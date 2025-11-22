#!/usr/bin/env python3
import requests
import json
from datetime import datetime

API_URL = "https://drapi.crwell.pro/api"

print("=" * 70)
print("TESTE COMPLETO - VISUALIZAÇÃO E DOWNLOAD DE EXAMES")
print("=" * 70)

# Login
print("\n1. Fazendo login...")
resp = requests.post(f"{API_URL}/auth/login", json={
    "email": "wasolutionscorp@gmail.com",
    "password": "senha123"
})

if resp.status_code != 200:
    print(f"✗ Erro no login: {resp.status_code}")
    exit(1)

TOKEN = resp.json()["token"]
print("✓ Login realizado com sucesso")

headers = {
    "Authorization": f"Bearer {TOKEN}",
    "Content-Type": "application/json"
}

# Get exams
print("\n2. Buscando exames do paciente #1...")
resp = requests.get(f"{API_URL}/exams?patient_id=1", headers=headers)

if resp.status_code != 200:
    print(f"✗ Erro ao buscar exames: {resp.status_code}")
    exit(1)

exams = resp.json().get("exams", [])
print(f"✓ Encontrados {len(exams)} exames")

if not exams:
    print("\n⚠ Nenhum exame encontrado. Teste encerrado.")
    exit(0)

# Test each exam
print("\n3. Testando download de cada exame:")
print("-" * 70)

success_count = 0
error_count = 0
errors = []

for i, exam in enumerate(exams, 1):
    exam_id = exam["id"]
    exam_name = exam["name"]

    print(f"\nExame {i}/{len(exams)}: {exam_name} (ID: {exam_id})")

    # Get download URL
    resp = requests.get(f"{API_URL}/exams/{exam_id}/download", headers=headers)

    if resp.status_code != 200:
        print(f"  ✗ Erro ao obter URL de download: {resp.status_code}")
        error_count += 1
        errors.append(f"Exame #{exam_id}: Erro {resp.status_code} ao obter URL")
        continue

    download_url = resp.json().get("download_url")
    expires_in = resp.json().get("expires_in", 0)

    print(f"  ✓ URL de download gerada (expira em {expires_in}s)")
    print(f"  URL: {download_url[:80]}...")

    # Test URL access
    print(f"  Testando acesso à URL...")
    test_resp = requests.head(download_url, timeout=10)

    if test_resp.status_code == 200:
        content_length = test_resp.headers.get('Content-Length', 'Unknown')
        content_type = test_resp.headers.get('Content-Type', 'Unknown')
        print(f"  ✓ Arquivo acessível!")
        print(f"    - Tipo: {content_type}")
        print(f"    - Tamanho: {content_length} bytes")
        success_count += 1
    else:
        print(f"  ✗ ERRO ao acessar arquivo: HTTP {test_resp.status_code}")
        error_count += 1
        errors.append(f"Exame #{exam_id}: HTTP {test_resp.status_code} ao acessar arquivo")

# Summary
print("\n" + "=" * 70)
print("RESUMO DOS TESTES")
print("=" * 70)
print(f"✓ Exames acessíveis: {success_count}")
print(f"✗ Exames com erro: {error_count}")

if errors:
    print("\nERROS ENCONTRADOS:")
    for error in errors:
        print(f"  - {error}")
else:
    print("\n✓ TODOS OS EXAMES ESTÃO ACESSÍVEIS!")
    print("\n🎉 Sistema de exames funcionando perfeitamente!")

print("\nData do teste:", datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
print("=" * 70)
