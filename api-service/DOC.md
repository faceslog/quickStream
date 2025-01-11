# Documentation de l’API

## Sommaire

1. [POST /api/publish](#post-apipublish)  
2. [GET /api/videos](#get-apivideos)  
3. [GET /api/status/:uuid (optionnel pour l’async)](#get-apistatusuuid)  

---

## POST /api/publish

**But**  
Publier (ou envoyer) une nouvelle vidéo. Dans le flux asynchrone, cette route se contente de stocker rapidement le fichier et de renvoyer un `uuid` pour le suivi.

**URL**  
```
POST /api/publish
```

**Paramètres/Champs**

- **Form Data** :
  - `title` (type: `string`)  
    \- **Obligatoire**  
    \- Le titre de la vidéo.  
    \- Limite : 255 caractères max.
  - `file` (type: `file`)  
    \- **Obligatoire**  
    \- Le fichier vidéo à uploader.

**Exemple de requête (curl)**  
```bash
curl -X POST http://localhost:8080/api/publish \
  -F "title=Ma super vidéo" \
  -F "file=@/chemin/vers/video.mp4"
```

**Réponses**  
- **201 Created** (ou **202 Accepted** si mode asynchrone)  
  - En cas de succès :  
  ```json
  {
    "uuid": "<identifiant-unique>",
    "status": "processing",
    "message": "File accepted for processing. Check back later."
  }
  ```
- **400 Bad Request**  
  - Si `title` est vide ou que le fichier n’est pas présent/valide.  
  ```json
  {
    "error": "title is required"
  }
  ```
  ou 
  ```json
  {
    "error": "file is required"
  }
  ```
  ou 
  ```json
  {
    "error": "invalid file MIME type"
  }
  ```
- **413 Request Entity Too Large**  
  - Si la taille du fichier dépasse la limite (ex. `config.MAX_FILE_SIZE`).  
  ```json
  {
    "error": "file size exceeds limit"
  }
  ```
- **409 Conflict**  
  - Si la vidéo existe déjà (même hash).  
  ```json
  {
    "error": "video already exists"
  }
  ```
- **500 Internal Server Error**  
  - Si la sauvegarde échoue (I/O, DB, etc.).  
  ```json
  {
    "error": "Oops, something went wrong with that file."
  }
  ```

---

## GET /api/videos

**But**  
Récupérer la liste de toutes les vidéos stockées dans la base de données, classées par date d’upload décroissante.

**URL**  
```
GET /api/videos
```

**Paramètres**  
Aucun paramètre particulier dans l’URL ou le body.

**Réponses**  
- **200 OK** : Renvoie un tableau d’objets `Video`.  
  ```json
  [
    {
      "uuid": "9b26bb90-xxxx-xxxx-xxxx-aaaaaaaaaaaa",
      "title": "Ma super vidéo",
      "hash": "abcdef1234567890...",
      "format": "mp4",
      "uploadedAt": "2025-01-10 15:00:00",
      "uri": "http://localhost:8080/videos/9b26bb90-xxxx-xxxx-xxxx-aaaaaaaaaaaa.mp4"
    },
    ...
  ]
  ```
- **500 Internal Server Error**  
  - Si la requête DB ou la récupération des vidéos échoue.  
  ```json
  {
    "error": "failed to fetch videos"
  }
  ```

---

## GET /api/status/:uuid

*(Uniquement si vous avez implémenté la logique asynchrone et un suivi de statut.)*

**But**  
Obtenir le statut d’un job d’upload/traitement en cours ou récemment terminé. Quand vous faites un `POST /api/publish` asynchrone, vous recevez un `uuid`. Vous pouvez ensuite consulter ce statut avec cet endpoint.

**URL**  
```
GET /api/status/:uuid
```
- `:uuid` : l’identifiant unique renvoyé lors de l’upload.

**Réponse**  
- **200 OK**  
  ```json
  {
    "uuid": "9b26bb90-xxxx-xxxx-xxxx-aaaaaaaaaaaa",
    "status": "processing"
  }
  ```
  \- `status` peut être `"pending"`, `"processing"`, `"completed"`, `"error"`, etc. selon votre implémentation.
- **404 Not Found**  
  - Si l’`uuid` est inconnu du système (job non trouvé).  
  ```json
  {
    "error": "job not found"
  }
  ```

---

# Structure des données

## Vidéo (Video)
- **uuid** (string, format UUID) : identifiant unique généré.  
- **title** (string) : titre de la vidéo.  
- **hash** (string) : hash unique du fichier pour détecter les doublons.  
- **format** (string) : format/extension du fichier (ex. `mp4`).  
- **uploadedAt** (string, format datetime) : date d’upload.  
- **uri** (string) : lien direct pour accéder à la vidéo via `router.Static("/videos", config.VideosDir)`.