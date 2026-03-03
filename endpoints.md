# GoWA API — Referência Completa de Endpoints (74 endpoints)

> **Base URL**: `http://localhost:3000` (ou conforme `APP_HOST` + `APP_PORT`)
>
> **Base Path**: Se configurado `--base-path="/api"`, adicione o prefixo a todos os endpoints.
>
> **Autenticação**: Basic Auth opcional (`--basic-auth=user:pass`). Incluir header `Authorization: Basic <base64>`.
>
> **Device Scoping**: Endpoints operacionais exigem `X-Device-Id` header ou `?device_id=<id>` query param. Se apenas um device estiver registrado, é usado automaticamente.

---

## 📱 Device Management (sem X-Device-Id)

### 1. Listar dispositivos

```bash
curl -X GET "http://localhost:3000/devices" \
  -H "Content-Type: application/json"
```

### 2. Adicionar dispositivo

```bash
curl -X POST "http://localhost:3000/devices" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "meu-dispositivo"
  }'
```

### 3. Obter info de um dispositivo

```bash
curl -X GET "http://localhost:3000/devices/meu-dispositivo" \
  -H "Content-Type: application/json"
```

### 4. Remover dispositivo

```bash
curl -X DELETE "http://localhost:3000/devices/meu-dispositivo" \
  -H "Content-Type: application/json"
```

### 5. Login via QR Code (por device)

```bash
curl -X GET "http://localhost:3000/devices/meu-dispositivo/login" \
  -H "Content-Type: application/json"
```

### 6. Login com pairing code (por device)

```bash
curl -X POST "http://localhost:3000/devices/meu-dispositivo/login/code?phone=5511999999999" \
  -H "Content-Type: application/json"
```

### 7. Logout de dispositivo

```bash
curl -X POST "http://localhost:3000/devices/meu-dispositivo/logout" \
  -H "Content-Type: application/json"
```

### 8. Reconectar dispositivo

```bash
curl -X POST "http://localhost:3000/devices/meu-dispositivo/reconnect" \
  -H "Content-Type: application/json"
```

### 9. Status do dispositivo

```bash
curl -X GET "http://localhost:3000/devices/meu-dispositivo/status" \
  -H "Content-Type: application/json"
```

---

## 🔐 App (requer X-Device-Id)

### 10. Login via QR Code

```bash
curl -X GET "http://localhost:3000/app/login" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 11. Login com pairing code

```bash
curl -X GET "http://localhost:3000/app/login-with-code?phone=5511999999999" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 12. Logout

```bash
curl -X GET "http://localhost:3000/app/logout" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 13. Reconectar

```bash
curl -X GET "http://localhost:3000/app/reconnect" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 14. Listar dispositivos conectados

```bash
curl -X GET "http://localhost:3000/app/devices" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 15. Status da conexão

```bash
curl -X GET "http://localhost:3000/app/status" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

---

## 💬 Send — Envio de Mensagens (requer X-Device-Id)

### 16. Enviar mensagem de texto

```bash
curl -X POST "http://localhost:3000/send/message" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "message": "Olá, tudo bem?",
    "reply_message_id": "",
    "mentions": [],
    "is_forwarded": false,
    "duration": null
  }'
```

#### Enviar com resposta a mensagem

```bash
curl -X POST "http://localhost:3000/send/message" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "message": "Respondendo sua mensagem",
    "reply_message_id": "3EB0A0B0C1D2E3F4"
  }'
```

#### Enviar com ghost mentions (menção sem @ no texto)

```bash
curl -X POST "http://localhost:3000/send/message" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511888888888@g.us",
    "message": "Atenção todos do grupo!",
    "mentions": ["5511999999999", "5511888888888"]
  }'
```

#### Menção @everyone (todos do grupo)

```bash
curl -X POST "http://localhost:3000/send/message" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511888888888@g.us",
    "message": "Aviso importante para todos!",
    "mentions": ["@everyone"]
  }'
```

### 17. Enviar imagem

```bash
curl -X POST "http://localhost:3000/send/image" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "caption=Olha essa foto!" \
  -F "image=@/caminho/para/imagem.jpg" \
  -F "compress=true" \
  -F "view_once=false"
```

#### Enviar imagem via URL

```bash
curl -X POST "http://localhost:3000/send/image" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "caption=Imagem da internet" \
  -F "image_url=https://exemplo.com/foto.jpg" \
  -F "compress=true" \
  -F "view_once=false"
```

### 18. Enviar arquivo/documento

```bash
curl -X POST "http://localhost:3000/send/file" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "caption=Segue o documento" \
  -F "file=@/caminho/para/documento.pdf"
```

#### Enviar arquivo via URL

```bash
curl -X POST "http://localhost:3000/send/file" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "caption=Relatório mensal" \
  -F "file_url=https://exemplo.com/relatorio.pdf"
```

### 19. Enviar vídeo

```bash
curl -X POST "http://localhost:3000/send/video" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "caption=Confira o vídeo!" \
  -F "video=@/caminho/para/video.mp4" \
  -F "compress=true" \
  -F "view_once=false"
```

#### Enviar vídeo via URL

```bash
curl -X POST "http://localhost:3000/send/video" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "caption=Vídeo da internet" \
  -F "video_url=https://exemplo.com/video.mp4" \
  -F "view_once=false"
```

### 20. Enviar áudio / nota de voz

```bash
curl -X POST "http://localhost:3000/send/audio" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "audio=@/caminho/para/audio.mp3" \
  -F "ptt=true"
```

> **`ptt=true`** = push-to-talk (nota de voz com botão azul). **`ptt=false`** = arquivo de áudio normal.

#### Enviar áudio via URL

```bash
curl -X POST "http://localhost:3000/send/audio" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "audio_url=https://exemplo.com/audio.ogg" \
  -F "ptt=true"
```

### 21. Enviar sticker

```bash
curl -X POST "http://localhost:3000/send/sticker" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "sticker=@/caminho/para/sticker.webp"
```

#### Enviar sticker via URL

```bash
curl -X POST "http://localhost:3000/send/sticker" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "phone=5511999999999" \
  -F "sticker_url=https://exemplo.com/sticker.webp"
```

> Aceita JPG, PNG, WebP, GIF. Converte automaticamente para WebP 512x512.

### 22. Enviar contato

```bash
curl -X POST "http://localhost:3000/send/contact" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "contact_name": "João Silva",
    "contact_phone": "5511888888888"
  }'
```

### 23. Enviar link com preview

```bash
curl -X POST "http://localhost:3000/send/link" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "link": "https://github.com/aldinokemal/go-whatsapp-web-multidevice",
    "caption": "Confira esse projeto!"
  }'
```

### 24. Enviar localização

```bash
curl -X POST "http://localhost:3000/send/location" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "latitude": "-23.5505",
    "longitude": "-46.6333"
  }'
```

### 25. Enviar enquete (poll)

```bash
curl -X POST "http://localhost:3000/send/poll" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "question": "Qual sua linguagem favorita?",
    "options": ["Go", "Python", "JavaScript", "Rust"],
    "max_answer": 1
  }'
```

### 26. Enviar presença global (online/offline)

```bash
curl -X POST "http://localhost:3000/send/presence" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "type": "available"
  }'
```

> `type`: `"available"` ou `"unavailable"`.

### 27. Enviar presença no chat (digitando/gravando)

```bash
curl -X POST "http://localhost:3000/send/chat-presence" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "action": "start"
  }'
```

> `action`: `"start"` (typing), `"stop"` (paused), `"recording"` (gravando áudio).

---

## 💬 Chat — Gerenciamento de Conversas (requer X-Device-Id)

### 28. Listar chats armazenados

```bash
curl -X GET "http://localhost:3000/chats?limit=25&offset=0&search=João" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

| Query Param | Tipo | Default | Descrição |
|-------------|------|---------|-----------|
| `limit` | int | 25 | Máximo de resultados |
| `offset` | int | 0 | Posição de início |
| `search` | string | — | Filtrar por nome |
| `has_media` | bool | false | Somente chats com mídia |

### 29. Obter mensagens de um chat

```bash
curl -X GET "http://localhost:3000/chat/5511999999999@s.whatsapp.net/messages?limit=50&offset=0" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

| Query Param | Tipo | Default | Descrição |
|-------------|------|---------|-----------|
| `limit` | int | 50 | Máximo de mensagens |
| `offset` | int | 0 | Posição de início |
| `media_only` | bool | false | Somente mensagens com mídia |
| `search` | string | — | Buscar texto nas mensagens |
| `start_time` | string | — | Data início (RFC3339: `2025-01-01T00:00:00Z`) |
| `end_time` | string | — | Data fim (RFC3339) |
| `is_from_me` | bool | — | Filtrar por mensagens enviadas/recebidas |

### 30. Fixar/desafixar chat

```bash
curl -X POST "http://localhost:3000/chat/5511999999999@s.whatsapp.net/pin" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "pinned": true
  }'
```

### 31. Configurar mensagens temporárias

```bash
curl -X POST "http://localhost:3000/chat/5511999999999@s.whatsapp.net/disappearing" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "timer_seconds": 86400
  }'
```

> Valores permitidos: `0` (desativado), `86400` (24h), `604800` (7 dias), `7776000` (90 dias).

### 32. Arquivar/desarquivar chat

```bash
curl -X POST "http://localhost:3000/chat/5511999999999@s.whatsapp.net/archive" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "archived": true
  }'
```

---

## ✉️ Message — Ações em Mensagens (requer X-Device-Id)

### 33. Reagir a uma mensagem

```bash
curl -X POST "http://localhost:3000/message/3EB0A0B0C1D2E3F4/reaction" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "emoji": "👍"
  }'
```

> Para remover reação, envie `"emoji": ""`.

### 34. Revogar (unsend) mensagem

```bash
curl -X POST "http://localhost:3000/message/3EB0A0B0C1D2E3F4/revoke" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999"
  }'
```

### 35. Deletar mensagem (local)

```bash
curl -X POST "http://localhost:3000/message/3EB0A0B0C1D2E3F4/delete" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999"
  }'
```

### 36. Editar mensagem

```bash
curl -X POST "http://localhost:3000/message/3EB0A0B0C1D2E3F4/update" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999",
    "message": "Texto atualizado da mensagem"
  }'
```

### 37. Marcar como lida

```bash
curl -X POST "http://localhost:3000/message/3EB0A0B0C1D2E3F4/read" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999"
  }'
```

### 38. Favoritar mensagem

```bash
curl -X POST "http://localhost:3000/message/3EB0A0B0C1D2E3F4/star" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999"
  }'
```

### 39. Desfavoritar mensagem

```bash
curl -X POST "http://localhost:3000/message/3EB0A0B0C1D2E3F4/unstar" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "phone": "5511999999999"
  }'
```

### 40. Download de mídia da mensagem

```bash
curl -X GET "http://localhost:3000/message/3EB0A0B0C1D2E3F4/download?phone=5511999999999" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

---

## 👥 Group — Gerenciamento de Grupos (requer X-Device-Id)

### 41. Criar grupo

```bash
curl -X POST "http://localhost:3000/group" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "title": "Meu Grupo Novo",
    "participants": ["5511999999999", "5511888888888"]
  }'
```

### 42. Entrar no grupo via link

```bash
curl -X POST "http://localhost:3000/group/join-with-link" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "link": "https://chat.whatsapp.com/ABCDEF123456"
  }'
```

### 43. Info do grupo via link de convite

```bash
curl -X GET "http://localhost:3000/group/info-from-link?link=https://chat.whatsapp.com/ABCDEF123456" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 44. Info do grupo por JID

```bash
curl -X GET "http://localhost:3000/group/info?group_id=120363012345678901@g.us" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 45. Sair do grupo

```bash
curl -X POST "http://localhost:3000/group/leave" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us"
  }'
```

### 46. Listar participantes do grupo

```bash
curl -X GET "http://localhost:3000/group/participants?group_id=120363012345678901@g.us" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 47. Exportar participantes (CSV)

```bash
curl -X GET "http://localhost:3000/group/participants/export?group_id=120363012345678901@g.us" \
  -H "X-Device-Id: meu-dispositivo" \
  -o participantes.csv
```

### 48. Adicionar participantes ao grupo

```bash
curl -X POST "http://localhost:3000/group/participants" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "participants": ["5511999999999", "5511888888888"]
  }'
```

### 49. Remover participantes do grupo

```bash
curl -X POST "http://localhost:3000/group/participants/remove" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "participants": ["5511999999999"]
  }'
```

### 50. Promover a admin

```bash
curl -X POST "http://localhost:3000/group/participants/promote" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "participants": ["5511999999999"]
  }'
```

### 51. Rebaixar de admin

```bash
curl -X POST "http://localhost:3000/group/participants/demote" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "participants": ["5511999999999"]
  }'
```

### 52. Listar solicitações de entrada

```bash
curl -X GET "http://localhost:3000/group/participant-requests?group_id=120363012345678901@g.us" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 53. Aprovar solicitações de entrada

```bash
curl -X POST "http://localhost:3000/group/participant-requests/approve" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "participants": ["5511999999999"]
  }'
```

### 54. Rejeitar solicitações de entrada

```bash
curl -X POST "http://localhost:3000/group/participant-requests/reject" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "participants": ["5511999999999"]
  }'
```

### 55. Definir foto do grupo

```bash
curl -X POST "http://localhost:3000/group/photo" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "group_id=120363012345678901@g.us" \
  -F "photo=@/caminho/para/foto.jpg"
```

### 56. Alterar nome do grupo

```bash
curl -X POST "http://localhost:3000/group/name" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "name": "Novo Nome do Grupo"
  }'
```

### 57. Bloquear/desbloquear configurações do grupo

```bash
curl -X POST "http://localhost:3000/group/locked" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "locked": true
  }'
```

### 58. Ativar/desativar modo anúncio (só admins enviam)

```bash
curl -X POST "http://localhost:3000/group/announce" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "announce": true
  }'
```

### 59. Definir descrição/tópico do grupo

```bash
curl -X POST "http://localhost:3000/group/topic" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "group_id": "120363012345678901@g.us",
    "topic": "Grupo para discutir projetos em Go"
  }'
```

### 60. Obter link de convite do grupo

```bash
curl -X GET "http://localhost:3000/group/invite-link?group_id=120363012345678901@g.us" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

#### Resetar link de convite

```bash
curl -X GET "http://localhost:3000/group/invite-link?group_id=120363012345678901@g.us&reset=true" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

---

## 👤 User — Informações de Usuário (requer X-Device-Id)

### 61. Info do contato

```bash
curl -X GET "http://localhost:3000/user/info?phone=5511999999999" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 62. Avatar do contato

```bash
curl -X GET "http://localhost:3000/user/avatar?phone=5511999999999&is_preview=false&is_community=false" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 63. Alterar seu avatar

```bash
curl -X POST "http://localhost:3000/user/avatar" \
  -H "X-Device-Id: meu-dispositivo" \
  -F "avatar=@/caminho/para/foto.jpg"
```

### 64. Alterar seu nome de exibição

```bash
curl -X POST "http://localhost:3000/user/pushname" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "push_name": "Meu Novo Nome"
  }'
```

### 65. Suas configurações de privacidade

```bash
curl -X GET "http://localhost:3000/user/my/privacy" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 66. Seus grupos

```bash
curl -X GET "http://localhost:3000/user/my/groups" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 67. Suas newsletters (canais)

```bash
curl -X GET "http://localhost:3000/user/my/newsletters" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 68. Seus contatos

```bash
curl -X GET "http://localhost:3000/user/my/contacts" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 69. Verificar se número está no WhatsApp

```bash
curl -X GET "http://localhost:3000/user/check?phone=5511999999999" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

### 70. Perfil comercial de um número

```bash
curl -X GET "http://localhost:3000/user/business-profile?phone=5511999999999" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo"
```

---

## 📰 Newsletter (requer X-Device-Id)

### 71. Deixar de seguir newsletter

```bash
curl -X POST "http://localhost:3000/newsletter/unfollow" \
  -H "Content-Type: application/json" \
  -H "X-Device-Id: meu-dispositivo" \
  -d '{
    "newsletter_id": "120363123456789012@newsletter"
  }'
```

---

## 🔗 Chatwoot Integration

### 72. Webhook do Chatwoot (sem autenticação)

> Este endpoint é chamado **pelo Chatwoot** automaticamente. Configure a URL de webhook no Chatwoot como:
> `http://seu-servidor:3000/chatwoot/webhook`

```bash
curl -X POST "http://localhost:3000/chatwoot/webhook" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message_created",
    "message_type": "outgoing",
    "content": "Mensagem do agente",
    "conversation": {
      "id": 123,
      "contact": {
        "phone_number": "+5511999999999"
      }
    }
  }'
```

### 73. Sincronizar histórico para Chatwoot

```bash
curl -X POST "http://localhost:3000/chatwoot/sync" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "meu-dispositivo",
    "days_limit": 3,
    "include_media": true,
    "include_groups": false
  }'
```

### 74. Status da sincronização Chatwoot

```bash
curl -X GET "http://localhost:3000/chatwoot/sync/status?device_id=meu-dispositivo" \
  -H "Content-Type: application/json"
```

---

## 📋 Resumo Rápido

| Categoria | Endpoints | X-Device-Id | Autenticação |
|-----------|-----------|-------------|--------------|
| Device Management | 9 | ❌ Não | Basic Auth (se configurado) |
| App | 6 | ✅ Sim | Basic Auth (se configurado) |
| Send | 12 | ✅ Sim | Basic Auth (se configurado) |
| Chat | 5 | ✅ Sim | Basic Auth (se configurado) |
| Message | 8 | ✅ Sim | Basic Auth (se configurado) |
| Group | 20 | ✅ Sim | Basic Auth (se configurado) |
| User | 10 | ✅ Sim | Basic Auth (se configurado) |
| Newsletter | 1 | ✅ Sim | Basic Auth (se configurado) |
| Chatwoot | 3 | ❌ Não | Webhook: sem auth / Sync: com auth |
| **Total** | **74** | | |

---

## 🔧 Notas para Integração com n8n

### Configuração do HTTP Request Node

1. **Base URL**: Configure como `http://seu-servidor:3000`
2. **Authentication**: Se Basic Auth estiver habilitado, use "Generic Credential Type" → "Basic Auth" com user/password
3. **Headers**: Adicione `X-Device-Id` como header fixo em todas as chamadas operacionais
4. **Para envio de arquivos**: Use "Body Content Type" = "Multipart Form Data"
5. **Para JSON**: Use "Body Content Type" = "JSON"

### Exemplo n8n — Enviar mensagem de texto

```
Method: POST
URL: http://localhost:3000/send/message
Authentication: Basic Auth (se habilitado)
Headers:
  X-Device-Id: meu-dispositivo
  Content-Type: application/json
Body (JSON):
{
  "phone": "{{ $json.phone }}",
  "message": "{{ $json.message }}"
}
```

### Exemplo n8n — Enviar imagem com arquivo do workflow

```
Method: POST
URL: http://localhost:3000/send/image
Authentication: Basic Auth (se habilitado)
Headers:
  X-Device-Id: meu-dispositivo
Body (Form Data):
  phone: {{ $json.phone }}
  caption: {{ $json.caption }}
  image: (binary data do node anterior)
```

### Exemplo n8n — Receber mensagens via Webhook

Configure um **Webhook Trigger** node no n8n e use a URL dele como `--webhook` flag:

```bash
./whatsapp rest --webhook="https://seu-n8n.com/webhook/whatsapp-events"
```

O payload recebido terá o formato:

```json
{
  "event": "message",
  "device_id": "5511999999999@s.whatsapp.net",
  "payload": {
    "id": "3EB0A0B0C1D2E3F4",
    "timestamp": "2025-01-01T12:00:00Z",
    "is_from_me": false,
    "from": "5511888888888@s.whatsapp.net",
    "from_name": "João",
    "chat_jid": "5511888888888@s.whatsapp.net",
    "message": "Olá, preciso de ajuda!",
    "media_type": "",
    "has_media": false
  }
}
```

**Header de segurança**: `X-Hub-Signature-256: sha256=<hmac>` (validar com `--webhook-secret`)

### Eventos disponíveis no webhook

| Evento | Descrição |
|--------|-----------|
| `message` | Nova mensagem recebida |
| `message.reaction` | Reação a mensagem |
| `message.revoked` | Mensagem revogada |
| `message.edited` | Mensagem editada |
| `message.ack` | Confirmação de leitura/entrega |
| `message.delete` | Mensagem deletada |
| `call.offer` | Chamada recebida |
| `group.participants` | Alteração em participantes |
| `group.info` | Info do grupo alterada |
| `presence` | Status de presença |
| `newsletter.join` | Entrou em newsletter |
| `newsletter.leave` | Saiu de newsletter |

> Filtrar eventos: `--webhook-events="message,message.ack,call.offer"`
