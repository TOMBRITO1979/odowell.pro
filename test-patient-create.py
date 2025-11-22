#!/usr/bin/env python3
import requests
import json
from datetime import datetime

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

# Test create patient
new_patient = {
    "name": f"Paciente Teste {datetime.now().strftime('%H%M%S')}",
    "email": f"teste{datetime.now().strftime('%H%M%S')}@teste.com",
    "phone": "(11) 99999-9999",
    "cpf": "000.000.000-00",
    "birth_date": "1990-01-01",
    "active": True
}

print("Testando criação de paciente...")
print(f"Dados: {json.dumps(new_patient, indent=2)}")
resp = requests.post(f"{API_URL}/patients", json=new_patient, headers=headers)
print(f"Status: {resp.status_code}")
print(f"Response: {json.dumps(resp.json(), indent=2)}")
