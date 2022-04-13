# Backlog
- better directory structure to localization
- better shared code amqp rabbit
- tests
- final readme with gif/video demo

# Known bugs and limitations
- Sometimes race condition if you try to record a very short audio.
- Cant really scale horizontally; There is an internal state using channels to manage the recording, cant handle the start and end of recording in different instances.