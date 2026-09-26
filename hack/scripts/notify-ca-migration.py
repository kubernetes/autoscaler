#!/usr/bin/env python3

# Copyright The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
Notify authors of open kubernetes/autoscaler pull requests that touch code
migrated to kubernetes-sigs/cluster-autoscaler.

The set of migrated paths is derived from an already-merged migration PR
(default: #10002). A path is considered "removed" when either:

  * it is a directory that was fully deleted by the migration PR, in which case
    every file under it matches (including files a stale PR adds that never
    existed on master), or
  * it is an individual deleted file living in a directory that still exists,
    in which case only that exact path matches.

Files that the migration PR merely *modified* (e.g. import-path rewrites in the
cloud providers that stayed behind) never match: those PRs still belong here and
only need a rebase, which GitHub and Prow already flag.

The script is read-only unless --execute is passed.

Examples:

  # See what would happen, no writes.
  hack/scripts/notify-ca-migration.py

  # Inspect the derived path rules only.
  hack/scripts/notify-ca-migration.py --print-rules

  # Post for real, one PR first.
  hack/scripts/notify-ca-migration.py --only-pr 9876 --execute

  # Post for real, everything, no prompt.
  hack/scripts/notify-ca-migration.py --execute --yes
"""

import argparse
import json
import os
import subprocess
import sys
import time

DEFAULT_REPO = "kubernetes/autoscaler"
DEFAULT_MIGRATION_PR = 10002
DEFAULT_ACTOR = "jackfrancis"

# Every posted comment ends with this marker so re-runs can skip PRs that were
# already notified.
MARKER = "<!-- ca-core-migration-notice -->"

DEFAULT_TEMPLATE = """\
@{author}, thanks for your contribution!

The Cluster Autoscaler core has moved out of `kubernetes/autoscaler` and now lives in
[kubernetes-sigs/cluster-autoscaler](https://github.com/kubernetes-sigs/cluster-autoscaler)
(see {repo}#{migration_pr}). This pull request changes files that no longer exist in this
repository, so it cannot be merged here and must be manually migrated to the new repo.

More information: https://groups.google.com/g/kubernetes-sig-autoscaling/c/rR8hJMMSGeI

Migrated paths touched by this PR:

{paths}
"""


def die(msg):
    print("error: %s" % msg, file=sys.stderr)
    sys.exit(1)


def git(args, cwd):
    proc = subprocess.run(["git"] + args, cwd=cwd, capture_output=True, text=True)
    if proc.returncode != 0:
        die("git %s failed: %s" % (" ".join(args), proc.stderr.strip()))
    return proc.stdout


def gh_api(path, method="GET", fields=None, accept_status=()):
    """Call the GitHub API through the gh CLI so auth is handled by gh."""
    cmd = ["gh", "api", "-X", method, "-H", "Accept: application/vnd.github+json", path]
    for key, value in (fields or {}).items():
        cmd += ["-f", "%s=%s" % (key, value)]
    proc = subprocess.run(cmd, capture_output=True, text=True)
    if proc.returncode != 0:
        stderr = proc.stderr.strip()
        for status in accept_status:
            if str(status) in stderr:
                return None, stderr
        die("gh api %s %s failed: %s" % (method, path, stderr))
    return json.loads(proc.stdout or "null"), ""


def gh_api_paginated(path, per_page=100, max_pages=100):
    items = []
    sep = "&" if "?" in path else "?"
    for page in range(1, max_pages + 1):
        batch, _ = gh_api("%s%sper_page=%d&page=%d" % (path, sep, per_page, page))
        if not batch:
            break
        items.extend(batch)
        if len(batch) < per_page:
            break
    return items


def resolve_migration_commit(repo_root, pr_number, explicit_commit):
    if explicit_commit:
        return explicit_commit.strip()
    out = git(
        ["log", "--format=%H", "--grep", "Merge pull request #%d " % pr_number, "-1"],
        repo_root,
    ).strip()
    if not out:
        die(
            "could not find the merge commit for PR #%d in the local clone; "
            "fetch the default branch or pass --migration-commit" % pr_number
        )
    return out


def derive_path_rules(repo_root, commit):
    """Return (removed_dirs, removed_files) for a merge commit."""
    name_status = git(
        ["diff", "--name-status", "-M", "%s^1" % commit, commit], repo_root
    )
    deleted = []
    for line in name_status.splitlines():
        parts = line.split("\t")
        if len(parts) < 2:
            continue
        status, path = parts[0], parts[-1]
        if status.startswith("D"):
            deleted.append(path)

    tree = git(["ls-tree", "-r", "--name-only", commit], repo_root).splitlines()
    surviving_dirs = set()
    for path in tree:
        segments = path.split("/")
        for i in range(1, len(segments)):
            surviving_dirs.add("/".join(segments[:i]))

    removed_dirs, removed_files = set(), set()
    for path in deleted:
        segments = path.split("/")
        prefix = None
        # Shortest ancestor directory that no longer exists post-merge.
        for i in range(1, len(segments)):
            candidate = "/".join(segments[:i])
            if candidate not in surviving_dirs:
                prefix = candidate
                break
        if prefix:
            removed_dirs.add(prefix)
        else:
            removed_files.add(path)
    return removed_dirs, removed_files


def migrated_paths(pr_files, removed_dirs, removed_files):
    return sorted(
        path
        for path in pr_files
        if path in removed_files
        or any(path == d or path.startswith(d + "/") for d in removed_dirs)
    )


def already_notified(repo, number):
    comments = gh_api_paginated("/repos/%s/issues/%d/comments" % (repo, number))
    return any(MARKER in (c.get("body") or "") for c in comments)


def render(template, pr, paths, repo, migration_pr, max_paths):
    shown = paths[:max_paths]
    listing = "\n".join("- `%s`" % p for p in shown)
    if len(paths) > len(shown):
        listing += "\n- ...and %d more" % (len(paths) - len(shown))
    body = template.format(
        author=(pr.get("user") or {}).get("login", "there"),
        pr=pr["number"],
        title=pr.get("title", ""),
        repo=repo,
        migration_pr=migration_pr,
        paths=listing,
    )
    return body.rstrip() + "\n\n" + MARKER


def post_comment(repo, number, body, sleep_seconds, retries=3):
    path = "/repos/%s/issues/%d/comments" % (repo, number)
    for attempt in range(1, retries + 1):
        result, stderr = gh_api(
            path, method="POST", fields={"body": body}, accept_status=(403, 429, 502)
        )
        if result is not None:
            time.sleep(sleep_seconds)
            return True
        backoff = sleep_seconds * (2 ** attempt) + 30
        print(
            "  ! post failed (attempt %d/%d), backing off %.0fs: %s"
            % (attempt, retries, backoff, stderr.splitlines()[0] if stderr else "")
        )
        if attempt == retries:
            return False
        time.sleep(backoff)
    return False


def parse_args():
    p = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    p.add_argument("--repo", default=DEFAULT_REPO)
    p.add_argument("--repo-root", default=os.path.join(os.path.dirname(__file__), "..", ".."))
    p.add_argument("--migration-pr", type=int, default=DEFAULT_MIGRATION_PR)
    p.add_argument(
        "--migration-commit",
        default="",
        help="merge commit of the migration PR; looked up in the local clone by default",
    )
    p.add_argument(
        "--base",
        default="master",
        help="only consider PRs targeting this base branch ('' for any)",
    )
    p.add_argument("--message-file", default=os.path.join(os.path.dirname(__file__), "ca-migration-notice.md"))
    p.add_argument("--max-paths", type=int, default=15, help="paths listed in the comment")
    p.add_argument("--only-pr", type=int, action="append", default=[])
    p.add_argument("--limit", type=int, default=0, help="stop after N matching PRs")
    p.add_argument("--include-bots", action="store_true")
    p.add_argument("--include-drafts", action="store_true", default=True)
    p.add_argument("--skip-drafts", dest="include_drafts", action="store_false")
    p.add_argument("--as", dest="actor", default=DEFAULT_ACTOR, help="expected gh login")
    p.add_argument("--sleep", type=float, default=3.0, help="seconds between comments")
    p.add_argument("--report", default="", help="write a JSON report to this path")
    p.add_argument("--print-rules", action="store_true", help="print path rules and exit")
    p.add_argument("--execute", action="store_true", help="actually post comments")
    p.add_argument("--yes", action="store_true", help="skip the confirmation prompt")
    return p.parse_args()


def main():
    args = parse_args()
    repo_root = os.path.abspath(args.repo_root)

    commit = resolve_migration_commit(repo_root, args.migration_pr, args.migration_commit)
    removed_dirs, removed_files = derive_path_rules(repo_root, commit)
    print(
        "migration commit %s: %d removed dirs, %d removed files"
        % (commit[:12], len(removed_dirs), len(removed_files))
    )
    if args.print_rules:
        print("\nremoved directories (prefix match):")
        for d in sorted(removed_dirs):
            print("  %s/" % d)
        print("\nremoved files (exact match):")
        for f in sorted(removed_files):
            print("  %s" % f)
        return

    template = DEFAULT_TEMPLATE
    if os.path.exists(args.message_file):
        with open(args.message_file) as fh:
            template = fh.read()

    if args.execute:
        me, _ = gh_api("/user")
        login = (me or {}).get("login")
        if args.actor and login != args.actor:
            die(
                "gh is authenticated as %r but --as expects %r; run "
                "`gh auth switch` or pass --as %s" % (login, args.actor, login)
            )
        print("authenticated as %s" % login)

    if args.only_pr:
        prs = []
        for number in args.only_pr:
            pr, _ = gh_api("/repos/%s/pulls/%d" % (args.repo, number))
            prs.append(pr)
    else:
        query = "/repos/%s/pulls?state=open" % args.repo
        if args.base:
            query += "&base=%s" % args.base
        prs = gh_api_paginated(query)
    print("scanning %d open PR(s)\n" % len(prs))

    matched, skipped_bots, already, failed = [], [], [], []
    for pr in prs:
        number = pr["number"]
        author = (pr.get("user") or {}).get("login", "ghost")
        if pr.get("draft") and not args.include_drafts:
            continue
        files = [
            f["filename"]
            for f in gh_api_paginated("/repos/%s/pulls/%d/files" % (args.repo, number), max_pages=30)
        ]
        hits = migrated_paths(files, removed_dirs, removed_files)
        if not hits:
            continue
        if author.endswith("[bot]") and not args.include_bots:
            skipped_bots.append(number)
            continue
        if already_notified(args.repo, number):
            already.append(number)
            continue
        matched.append({"number": number, "author": author, "title": pr.get("title", ""), "paths": hits, "pr": pr})
        if args.limit and len(matched) >= args.limit:
            break

    print("%d PR(s) match migrated paths and are not yet notified" % len(matched))
    if already:
        print("%d PR(s) already carry the notice, skipped: %s" % (len(already), already))
    if skipped_bots:
        print("%d bot PR(s) skipped (--include-bots to include): %s" % (len(skipped_bots), skipped_bots))
    print()

    for entry in matched:
        print("#%-6d %-22s %s" % (entry["number"], "@" + entry["author"], entry["title"][:70]))
        for path in entry["paths"][:5]:
            print("            %s" % path)
        if len(entry["paths"]) > 5:
            print("            ...and %d more" % (len(entry["paths"]) - 5))

    if not args.execute:
        print("\nDRY RUN: nothing was posted. Re-run with --execute to comment.")
        if matched:
            print("\n--- sample comment for #%d ---" % matched[0]["number"])
            print(render(template, matched[0]["pr"], matched[0]["paths"], args.repo, args.migration_pr, args.max_paths))
    elif matched:
        if not args.yes:
            answer = input("\nPost %d comment(s) to %s as %s? [y/N] " % (len(matched), args.repo, args.actor))
            if answer.strip().lower() not in ("y", "yes"):
                die("aborted by user")
        print()
        for entry in matched:
            number = entry["number"]
            body = render(template, entry["pr"], entry["paths"], args.repo, args.migration_pr, args.max_paths)
            if post_comment(args.repo, number, body, args.sleep):
                print("#%-6d commented (@%s)" % (number, entry["author"]))
            else:
                print("#%-6d FAILED" % number)
                failed.append(number)
        print("\ndone: %d posted, %d already notified, %d failed"
              % (len(matched) - len(failed), len(already), len(failed)))

    if args.report:
        report = {
            "repo": args.repo,
            "migration_pr": args.migration_pr,
            "migration_commit": commit,
            "executed": args.execute,
            "matched": [{k: v for k, v in e.items() if k != "pr"} for e in matched],
            "skipped_bots": skipped_bots,
            "already_notified": already,
            "failed": failed,
        }
        with open(args.report, "w") as fh:
            json.dump(report, fh, indent=2)
        print("report written to %s" % args.report)

    if failed:
        sys.exit(1)


if __name__ == "__main__":
    main()
