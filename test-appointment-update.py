#!/usr/bin/env python3
import requests
import json

API_URL = "https://drapi.crwell.pro/api"

# Login
resp = requests.post(f"{API_URL}/auth/login", json={
    "email": "wasolutionscorp@gmail.com",
    "password": "senha123"
})
TOKEN = resp.json()["token"]

headers = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json"
}

# Test update
update_data = {
    "patient_id": 5,
    "dentist_id": 4,
    "start_time": "2025-10-04T01:59:00.655Z",
    "end_time": "2025-10-04T02:59:00.655Z",
    "type": "regular",
    "procedure": "whitening",
    "status": "completed",
    "notes": "TESTE MANUAL - FUNCIONOU?",
    "confirmed": False,
    "is_recurring": False,
    "recurrence_rule": ""
}

print("Testando UPDATE de appointment...")
resp = requests.put(f"{API_URL}/appointments/56", json=update_data, headers=headers)
print(f"Status: {resp.status_code}")
print(f"Response: {json.dumps(resp.json(), indent=2)}")
