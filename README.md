# Backlog
- tests
- final readme with gif/video demo
# Known bugs and limitations
- Sometimes race condition if you try to record a very short audio.
- Cant really scale horizontally; There is an internal state using channels to manage the recording, cant handle the start and end of recording in different instances.
- Not meant for unstable connections: if you are outside with the phone and trying to record an audio and you have low signal, you most likely will lose that audio.