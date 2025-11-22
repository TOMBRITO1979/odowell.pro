#!/usr/bin/env python3
import requests
import json
from datetime import datetime, timedelta

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

# Test create appointment
new_apt = {
    "patient_id": 5,
    "dentist_id": 4,
    "start_time": (datetime.now() + timedelta(days=7)).isoformat(),
    "end_time": (datetime.now() + timedelta(days=7, hours=1)).isoformat(),
    "type": "regular",
    "procedure": "Consulta de teste",
    "status": "scheduled",
    "notes": f"TESTE CRIADO - {datetime.now().strftime('%H:%M:%S')}"
}

print("Testando criação de agendamento...")
print(f"Dados: {json.dumps(new_apt, indent=2)}")
resp = requests.post(f"{API_URL}/appointments", json=new_apt, headers=headers)
print(f"Status: {resp.status_code}")
print(f"Response: {json.dumps(resp.json(), indent=2)}")
