#!/usr/bin/env bash
# MoneyFlow: ручные запросы (curl). Привязаны к тест-кейсам из 02-test-cases.md.
# Запуск: API на :8080, затем  bash postman/curl-examples.sh
API=${API_URL:-http://localhost:8080}
JSON='Content-Type: application/json'
id() { grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4; }

echo "== Подготовка: счета A (RUB 100000), B (RUB 0), U (USD 0)"
A=$(curl -s -X POST $API/api/v1/accounts -H "$JSON" -d '{"owner_name":"Ivan","currency":"RUB","initial_balance":100000}' | id)
B=$(curl -s -X POST $API/api/v1/accounts -H "$JSON" -d '{"owner_name":"Petr","currency":"RUB"}' | id)
U=$(curl -s -X POST $API/api/v1/accounts -H "$JSON" -d '{"owner_name":"Usd","currency":"USD"}' | id)
echo "A=$A B=$B U=$U"

echo "== TC-ACC-001: создание счёта -> 201"
curl -i -X POST $API/api/v1/accounts -H "$JSON" -d '{"owner_name":"Ivan","currency":"RUB","initial_balance":100000}'

echo "== TC-ACC-003: пустое имя -> 400 VALIDATION_ERROR"
curl -i -X POST $API/api/v1/accounts -H "$JSON" -d '{"owner_name":"","currency":"RUB"}'

echo "== TC-ACC-004: валюта XXX -> 400"
curl -i -X POST $API/api/v1/accounts -H "$JSON" -d '{"owner_name":"Ivan","currency":"XXX"}'

echo "== TC-ACC-008: получить счёт -> 200"
curl -i $API/api/v1/accounts/$A

echo "== Баланс -> 200"
curl -i $API/api/v1/accounts/$A/balance

echo "== TC-ACC-009: невалидный UUID -> 400 INVALID_ID"
curl -i $API/api/v1/accounts/not-a-uuid

echo "== TC-ACC-010: несуществующий счёт -> 404"
curl -i $API/api/v1/accounts/00000000-0000-4000-8000-000000000000

echo "== TC-TRF-001: перевод 5000 -> 201, A=95000, B=5000"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-1' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":5000}"

echo "== TC-TRF-014: повтор того же запроса -> 200 + Idempotent-Replay: true"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-1' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":5000}"

echo "== TC-TRF-015: тот же ключ, другая сумма -> 409"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-1' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":1}"

echo "== TC-TRF-002: разные валюты -> 422 CURRENCY_MISMATCH"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-2' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$U\",\"amount\":100}"

echo "== TC-TRF-003: недостаточно средств -> 422"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-3' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":999999999}"

echo "== TC-TRF-004/005: сумма -1 и 0 -> 400"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-4' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":-1}"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-5' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":0}"

echo "== TC-TRF-009: перевод на тот же счёт -> 400"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-6' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$A\",\"amount\":100}"

echo "== TC-TRF-013: без Idempotency-Key -> 400"
curl -i -X POST $API/api/v1/transfers -H "$JSON" \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":100}"

echo "== TC-ACC-011/012: блокировка B -> 200, повтор -> 409"
curl -i -X PATCH $API/api/v1/accounts/$B/block
curl -i -X PATCH $API/api/v1/accounts/$B/block

echo "== TC-TRF-012: перевод на заблокированный счёт -> 422 ACCOUNT_BLOCKED"
curl -i -X POST $API/api/v1/transfers -H "$JSON" -H 'Idempotency-Key: demo-7' \
  -d "{\"from_account_id\":\"$A\",\"to_account_id\":\"$B\",\"amount\":100}"

echo "== TC-TRF-016: история переводов -> 200"
curl -i "$API/api/v1/accounts/$A/transfers?limit=10&offset=0"

echo "== TC-TRF-017: limit=0 -> 400"
curl -i "$API/api/v1/accounts/$A/transfers?limit=0"