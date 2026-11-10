#!/bin/sh
set -eu

LOG_URL=${LOG_URL:-http://127.0.0.1:3905}
TOKEN=${LOG_ADMIN_TOKEN:-devtoken}

admin() {
  curl -sf -X "$1" "$LOG_URL$2" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' ${3:+-d "$3"}
}

steam_game() {
  admin POST /admin/items "{\"kind\":\"game\",\"title\":\"$1\",\"fields\":{\"steam_appid\":$2,\"cover\":\"https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/$2/header.jpg\",\"playtime_minutes\":$3,\"achievements_unlocked\":$4,\"achievements_total\":$5,\"steam_categories\":$6,\"genres\":$7}}" |
    python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])'
}

entry() {
  admin PATCH "/admin/items/$1/entry" "$2" > /dev/null
}

hades=$(steam_game "Hades" 1145360 3120 38 49 '["Single-player","Steam Trading Cards"]' '["Action","Indie","RPG"]')
disco=$(steam_game "Disco Elysium" 632470 2400 20 45 '["Single-player","Steam Trading Cards"]' '["RPG"]')
wilds=$(steam_game "Outer Wilds" 753640 900 0 31 '["Single-player"]' '["Adventure","Indie"]')
steam_game "Some Card Farm Game" 400 140 0 0 '["Single-player","Steam Trading Cards"]' '["Casual"]' > /dev/null
steam_game "Another Idled One" 620 95 0 12 '["Steam Trading Cards"]' '["Puzzle"]' > /dev/null
steam_game "Never Launched Bundle Game" 220 0 0 0 '["Single-player"]' '["Action"]' > /dev/null
steam_game "Half Hour Curiosity" 570 35 0 0 '["Multi-player"]' '["Action"]' > /dev/null
steam_game "Wallpaper Engine" 431960 40000 0 0 '["Steam Workshop"]' '["Casual","Utilities"]' > /dev/null

entry "$hades" '{"status":"finished","visibility":"public","scores":{"enjoyment":92,"quality":90},"blurb":"The loop that made me like roguelikes.","addTags":["favorite/all-time","genre/roguelike"]}'
entry "$disco" '{"status":"played","visibility":"public","scores":{"enjoyment":88},"blurb":"Reads like a novel that argues back.","addTags":["genre/rpg"]}'
entry "$wilds" '{"status":"finished","visibility":"private","scores":{"enjoyment":98},"notes":"private notes never reach public pages"}'
admin POST "/admin/items/$hades/moments" '{"at":"2024-03-02","note":"First clear."}' > /dev/null
admin POST "/admin/items/$hades/moments" '{"at":"2024-03-05","note":"private moment","visibility":"private"}' > /dev/null

thing=$(admin POST /admin/items '{"kind":"thing","title":"A Very Good Bench","fields":{"description":"In a park."}}' | python3 -c 'import json,sys;print(json.load(sys.stdin)["id"])')
entry "$thing" '{"status":"seen","visibility":"public","scores":{"enjoyment":75}}'

echo "seeded $LOG_URL"
