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

# Get exams for a patient (assuming patient ID 1)
print("Buscando exames do paciente #1...")
resp = requests.get(f"{API_URL}/exams?patient_id=1", headers=headers)
print(f"Status: {resp.status_code}")
if resp.status_code == 200:
    exams = resp.json().get("exams", [])
    print(f"Total de exames: {len(exams)}")

    if exams:
        # Try to get download URL for first exam
        exam_id = exams[0]["id"]
        print(f"\nTentando baixar exame #{exam_id}...")

        resp = requests.get(f"{API_URL}/exams/{exam_id}/download", headers=headers)
        print(f"Status: {resp.status_code}")
        print(f"Response: {json.dumps(resp.json(), indent=2)}")

        if resp.status_code == 200:
            download_url = resp.json().get("download_url")
            print(f"\nURL de download gerada: {download_url[:100]}...")

            # Try to access the URL
            print("\nTestando acesso à URL...")
            test_resp = requests.head(download_url)
            print(f"Status do acesso: {test_resp.status_code}")
            if test_resp.status_code != 200:
                print(f"Headers: {test_resp.headers}")
    else:
        print("Nenhum exame encontrado para este paciente")
else:
    print(f"Erro: {resp.json()}")
