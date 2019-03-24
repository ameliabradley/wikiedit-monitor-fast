# wikiedit-monitor-fast

This is an experiment at writing a faster and more reliable version of the [wikiedit-monitor](https://github.com/leebradley/wikiedit-monitor) project.

The Wikimedia API has updated substantially since I created that project with [Barbarrosa](https://github.com/Barbarrosa)
- The endpoint for bulk querying diffs is deprecated in favor of querying diffs indiviudally
- The Websocket API is deprecated in favor of [Server-sent events](https://en.wikipedia.org/wiki/Server-sent_events)
