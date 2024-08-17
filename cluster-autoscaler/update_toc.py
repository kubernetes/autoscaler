#!/usr/bin/env python3

# Copyright 2017 The Kubernetes Authors.
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

import re

SECTION_PREFIX = "# "
QUESTION_PREFIX = "### "

def updateFAQ():
    with open("FAQ.md","r") as faq_file:
        faq_content = faq_file.read()
    faq_lines = faq_content.splitlines()
    while faq_lines and faq_lines[-1] == '':
        faq_lines = faq_lines[:-1]

    prefixes = (SECTION_PREFIX, QUESTION_PREFIX)
    toc_elements = []
    after_toc = False
    for line in faq_lines:
        if line.strip() == "<!--- TOC END -->":
            after_toc = True
        if not after_toc:
            continue
        for i, pref in enumerate(prefixes):
            if line.strip().startswith(pref):
                processed_line = line.strip()[len(pref):].strip()
                if processed_line[-1] == ':':
                    processed_line = processed_line[:-1]
                toc_elements.append((processed_line, i))
    in_toc = False

    with open("FAQ.md","w") as faq_file:
        for line in faq_lines:
            if line.strip() == "<!--- TOC BEGIN -->":
                in_toc = True
                faq_file.write(line +"\n")
                for question, indent in toc_elements:
                    faq_file.write("%s* [%s](#%s)\n" % (' ' * 2 * indent, question, re.sub("[^a-z0-9\- ]+", "", question.lower()).replace(" ","-")))
            if line.strip() == "<!--- TOC END -->":
                in_toc = False
            if not in_toc:
                faq_file.write(line+"\n")

if __name__ == '__main__':
    updateFAQ()
