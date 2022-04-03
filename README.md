# Ideas
- commands. /tateru to repeat greeting. /tateru-stats for voice recording stats and ranking

# Backlog
- postgre database
- better directory structure to localization
- better shared code amqp rabbit
- tests
- final readme with gif/video demo

# Frozen
- progress of audio recording and status monitoring, avisa al usuario si el audio no se esta guardando bien...
- buses using generics if possible

# Known bugs and limitations
- Sometimes race condition if you try to record a very short audio.
- Cant really scale horizontally; There is an internal state using channels to manage the recording, cant handle the start and end of recording in different instances.