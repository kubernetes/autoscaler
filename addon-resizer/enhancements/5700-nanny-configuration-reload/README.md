# KEP-5546: Automatic reload of nanny configuration when updated

<!-- toc -->
- [Summary](#summary)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
    - [Notes](#notes)
    - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
    - [Test Plan](#test-plan)
<!-- /toc -->

Sure, here's the enhancement proposal in the requested format:

## Summary
- **Goals:** The goal of this enhancement is to improve the user experience for applying nanny configuration changes in the addon-resizer 1.8 when used with the metrics server. The proposed solution involves automatically reloading the nanny configuration whenever changes occur, eliminating the need for manual intervention and sidecar containers. 
- **Non-Goals:** This proposal does not aim to update the functional behavior of the addon-resizer.

## Proposal
The proposed solution involves updating the addon-resizer with the following steps:
- Create a file system watcher using `fsnotify` under `utils/fswatcher` to watch nanny configurations' changes. It should run as a goroutine in the background.
- Detect changes of the nanny configurations' file using the created `fswatcher` trigger the reloading process when configuration changes are detected. Events should be sent in a channel.
- Re-execute the method responsible for building the NannyConfiguration `loadNannyConfiguration` to apply the updated configuration to the addon-resizer.
- Proper error handling should be implemented to manage scenarios where the configuration file is temporarily inaccessible or if there are parsing errors in the configuration file.

### Risks and Mitigations
- There is a potential risk of filesystem-related issues causing the file watcher to malfunction. Proper testing and error handling should be implemented to handle such scenarios gracefully.
- Errors in the configuration file could lead to unexpected behavior or crashes. The addon-resizer should handle parsing errors and fall back to the previous working configuration if necessary.

## Design Details
- Create a new package for the `fswatcher` under `utils/fswatcher`. It would contain the `fswatcher` struct and methods and unit-tests.
    - `FsWatcher` struct would look similar to this:
    ```go
    type FsWatcher struct {
        *fsnotify.Watcher

        Events    chan struct{}
        ratelimit time.Duration
        names     []string
        paths     map[string]struct{}
    }
    ```
    - Implement the following functions:
        - `CreateFsWatcher`: Instantiates a new `FsWatcher` and start watching on file system.
        - `initWatcher`: Initializes the `fsnotify` watcher and initialize the `paths` that would be watched.
        - `add`: Adds a new file to watch.
        - `reset`: Re-initializes the `FsWatcher`.
        - `watch`: watches for the configured files.
- In the main function, we create a new `FsWatcher` and then we wait in an infinite loop to receive events indicating
filesystem changes. Based on these changes, we re-execute `loadNannyConfiguration` function.

> **Note:** The expected configuration file format is YAML. It has the same structure as the NannyConfiguration CRD.

### Test Plan
To ensure the proper functioning of the enhanced addon-resizer, the following test plan should be executed:
1. **Unit Tests:** Write unit tests to validate the file watcher's functionality and ensure it triggers events when the configuration file changes.
2. **Manual e2e Tests:** Deploy the addon-resizer with `BaseMemory` of `300Mi` and then we change the `BaseMemory` to `100Mi`. We should observer changes in the behavior of watched pod.


[fsnotify]: https://github.com/fsnotify/fsnotify
