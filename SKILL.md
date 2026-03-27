# xsh - Guide pour LLM

Guide de référence pour utiliser xsh (Twitter/X CLI) via Model Context Protocol (MCP) ou commandes directes.

## Qu'est-ce que xsh ?

CLI pour Twitter/X utilisant l'authentification par cookies de navigateur. Aucune clé API requise. Fonctionne en mode terminal (humain) ou JSON (LLM/AI).

## Installation Rapide

```bash
# Option 1: Via Go
go install github.com/benoitpetit/xsh@latest

# Option 2: Script Linux/macOS
curl -sSL https://raw.githubusercontent.com/benoitpetit/xsh/master/scripts/install.sh | bash

# Option 3: Binaire direct
wget https://github.com/benoitpetit/xsh/releases/latest/download/xsh-linux-amd64 -O xsh
chmod +x xsh && sudo mv xsh /usr/local/bin/
```

## Authentification (Obligatoire)

Avant toute commande, authentifiez-vous:

```bash
# Extraction auto des cookies du navigateur
xsh auth login

# Ou import depuis Cookie Editor
xsh auth import cookies.json

# Vérifier le statut
xsh auth status
```

**Multi-compte:**
```bash
xsh auth login --account work    # Créer compte nommé
xsh auth switch work             # Changer de compte
xsh auth accounts                # Lister les comptes
```

## Formats de Sortie (Important pour LLM)

Toutes les commandes supportent:

```bash
--json      # JSON complet
--compact   # JSON minimal (essentiel uniquement)
--yaml      # YAML
```

**Mode auto-détection:** Si stdout est redirigé (pipe), le JSON est automatiquement utilisé.

## Commandes par Usage

### 1. Timeline & Découverte

```bash
# Timeline personnelle
xsh feed --count 20 --json
xsh feed --type following --count 50

# Recherche
xsh search "golang" --type Latest --pages 2 --json
xsh search "python" --type Top --count 100

# Tendances
xsh trends --location "France"
xsh trends --woeid 615702  # Paris
```

### 2. Tweets

```bash
# Voir un tweet
xsh tweet view <tweet_id> --thread --json

# Publier
xsh tweet post "Hello World!"
xsh tweet post "Avec image" --image photo.jpg
xsh tweet post "Réponse" --reply-to <tweet_id>
xsh tweet post "Citation" --quote <tweet_url>

# Engagement
xsh tweet like <tweet_id>
xsh tweet retweet <tweet_id>
xsh tweet bookmark <tweet_id>

# Actions rapides (root level)
xsh unlike <tweet_id>
xsh unretweet <tweet_id>
xsh unbookmark <tweet_id>

# Supprimer
xsh tweet delete <tweet_id>
```

### 3. Utilisateurs

```bash
# Profil
xsh user <handle> --json

# Tweets d'un utilisateur
xsh user tweets <handle> --count 50 --json
xsh user tweets <handle> --replies  # Inclure réponses

# Likes
xsh user likes <handle> --count 50

# Réseau
xsh user followers <handle> --count 100
xsh user following <handle> --count 100

# Actions sociales
xsh follow <handle>
xsh unfollow <handle>
xsh block <handle>
xsh mute <handle>
```

### 4. Batch Operations (Multiple IDs)

```bash
# Récupérer plusieurs tweets
xsh tweets <id1> <id2> <id3> --json

# Récupérer plusieurs utilisateurs
xsh users <handle1> <handle2> <handle3> --json
```

### 5. Listes

```bash
# Lister mes listes
xsh lists --json

# Voir tweets d'une liste
xsh lists view <list_id> --count 50

# Gérer
xsh lists create "Ma Liste" --description "Description"
xsh lists add-member <list_id> <handle>
xsh lists remove-member <list_id> <handle>
xsh lists delete <list_id>
```

### 6. Bookmarks

```bash
xsh bookmarks --count 50 --json
xsh bookmarks-folders
xsh bookmarks-folder <folder_id>
```

### 7. Messages Directs

```bash
xsh dm inbox --json
xsh dm send <handle> "Message text"
xsh dm delete <message_id>
```

### 8. Tweets Programmés

```bash
xsh schedule "Futur tweet" --at "2026-04-01 09:00"
xsh scheduled --json
xsh unschedule <scheduled_tweet_id>
```

### 9. Jobs

```bash
xsh jobs search "software engineer" --location "Paris" --json
xsh jobs search "devops" --location-type remote --count 50
xsh jobs view <job_id>
```

### 10. Médias & Export

```bash
# Télécharger média d'un tweet
xsh download <tweet_id> --output-dir ./media

# Exporter en différents formats
xsh export feed --format json --output tweets.json
xsh export feed --format csv --output tweets.csv
xsh export search "golang" --format jsonl --output results.jsonl
xsh export bookmarks --format md --output bookmarks.md

# Formats supportés: json, jsonl, csv, tsv, md
```

### 11. Composition de Threads

```bash
# Mode interactif
xsh compose

# Depuis fichier
xsh compose --file thread.txt --dry-run
```

### 12. Utilitaires

```bash
# Compter caractères
xsh count "Texte du tweet"
xsh count --file draft.txt

# Diagnostics
xsh doctor --json
xsh status --json

# Endpoints (gestion interne)
xsh endpoints list
xsh endpoints refresh
xsh auto-update
```

## Schémas JSON

### Tweet

```json
{
  "id": "1234567890",
  "text": "Contenu du tweet",
  "author_id": "987654321",
  "author_name": "Nom",
  "author_handle": "username",
  "author_verified": true,
  "created_at": "2024-01-15T10:30:00Z",
  "engagement": {
    "likes": 42,
    "retweets": 12,
    "replies": 5,
    "views": 1500,
    "bookmarks": 8,
    "quotes": 2
  },
  "media": [
    {
      "type": "photo",
      "url": "https://..."
    }
  ],
  "is_retweet": false,
  "reply_to_id": "",
  "conversation_id": "1234567890"
}
```

### User

```json
{
  "id": "987654321",
  "handle": "username",
  "name": "Nom Complet",
  "bio": "Description...",
  "location": "Paris",
  "website": "https://...",
  "verified": true,
  "followers_count": 1000,
  "following_count": 500,
  "tweet_count": 5000,
  "created_at": "2020-01-01T00:00:00Z",
  "profile_image_url": "https://..."
}
```

### Timeline Response

```json
{
  "tweets": [...],
  "cursor_top": "...",
  "cursor_bottom": "...",
  "has_more": true
}
```

## MCP Server (Model Context Protocol)

Démarrer le serveur MCP:
```bash
xsh mcp
```

### Configuration Claude Desktop

`~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "xsh": {
      "command": "xsh",
      "args": ["mcp"]
    }
  }
}
```

### Outils MCP Disponibles (52)

**Lecture (24):**
- `get_feed`, `search`, `get_tweet`, `get_tweet_thread`
- `get_tweets_batch`, `get_users_batch`
- `get_user`, `get_user_tweets`, `get_user_likes`
- `get_followers`, `get_following`
- `list_bookmarks`, `get_bookmark_folders`, `get_bookmark_folder_timeline`
- `get_lists`, `get_list_timeline`, `get_list_members`
- `dm_inbox`, `get_trending`, `search_jobs`, `get_job`, `auth_status`

**Écriture (14):**
- `post_tweet`, `delete_tweet`
- `like`, `unlike`, `retweet`, `unretweet`, `bookmark`, `unbookmark`
- `follow`, `unfollow`, `block`, `unblock`, `mute`, `unmute`

**Admin (14):**
- `create_list`, `delete_list`, `add_list_member`, `remove_list_member`, `pin_list`, `unpin_list`
- `schedule_tweet`, `list_scheduled_tweets`, `cancel_scheduled_tweet`
- `dm_send`, `dm_delete`
- `download_media`

## Exemples pour LLM / Scripts

### Pipeline de Traitement

```bash
# Extraire textes des tweets
xsh feed --json | jq '.[].text'

# Filtrer par engagement
xsh search "golang" --json | jq '[.[] | select(.engagement.likes > 100)]'

# Analyse de followers
xsh user followers elonmusk --json | jq 'map(.followers_count) | add'
```

### Export pour Analyse

```bash
# Exporter timeline complète
xsh export feed --format json --output timeline.json

# Convertir en CSV pour Excel
xsh export search "keyword" --format csv --output results.csv
```

### Surveillance Continue

```bash
# Monitorer mentions
while true; do
  xsh search "@moncompte" --json | jq '.[].text'
  sleep 60
done
```

## Configuration

Fichier: `~/.config/xsh/config.toml`

```toml
default_count = 20

[display]
theme = "default"
show_engagement = true
max_width = 100

[request]
delay = 1.5
timeout = 30
max_retries = 3
```

## Codes de Sortie

| Code | Signification |
|------|---------------|
| 0 | Succès |
| 1 | Erreur générale |
| 2 | Erreur d'authentification |
| 3 | Rate limit |

## Résolution de Problèmes

```bash
# Vérifier santé du système
xsh doctor

# Rafraîchir endpoints
xsh endpoints refresh

# Mettre à jour endpoints obsolètes
xsh auto-update

# Mode verbose (debug)
xsh feed --verbose
```

## Bonnes Pratiques

1. **Rate Limiting**: Utilisez `--delay` ou configurez `delay` dans config.toml
2. **Batch**: Privilégiez `tweets` et `users` pour plusieurs IDs
3. **Pipes**: Les commandes détectent automatiquement les pipes et sortent en JSON
4. **Comptes**: Utilisez des comptes nommés pour gérer multiples profils
5. **Export**: Utilisez `export` pour sauvegarder des données en masse

## Sécurité

- Credentials stockés avec permissions 0600
- TLS fingerprinting (uTLS) pour éviter la détection bot
- Aucune donnée envoyée à des tiers
- Configuration locale uniquement
