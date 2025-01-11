# quickStream

Upload a video stream it for friends

- Suppression tout les X temps
- Lire les vidéos de X chunk en X chunk

```sh pdf
curl -X POST http://localhost:8080/api/publish \
     -F "file=@test.mp4" \
     -F "title=Test Title"
```