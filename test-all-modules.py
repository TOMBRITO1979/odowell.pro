#!/usr/bin/env python3
import requests
import json
from datetime import datetime, timedelta

API_URL = "https://drapi.crwell.pro/api"
TOKEN = None

# Contadores
edited_count = 0
created_count = 0
errors = []

def login():
    global TOKEN
    response = requests.post(f"{API_URL}/auth/login", json={
        "email": "wasolutionscorp@gmail.com",
        "password": "senha123"
    })
    if response.status_code == 200:
        TOKEN = response.json()["token"]
        print("✓ Login realizado com sucesso")
        return True
    else:
        print(f"✗ Falha no login: {response.status_code}")
        return False

def get_headers():
    return {
        "Authorization": f"Bearer {TOKEN}",
        "Content-Type": "application/json"
    }

def test_appointments():
    global edited_count, created_count, errors
    print("\n=== TESTANDO: Agendamentos ===")

    # Listar
    resp = requests.get(f"{API_URL}/appointments?page=1&page_size=10", headers=get_headers())
    if resp.status_code != 200:
        errors.append(f"Appointments GET: {resp.status_code}")
        return

    appointments = resp.json().get("appointments", [])
    if appointments:
        # Editar primeiro
        apt = appointments[0]
        apt_id = apt["id"]
        apt["notes"] = f"TESTE EDITADO - {datetime.now().strftime('%H:%M:%S')}"

        resp = requests.put(f"{API_URL}/appointments/{apt_id}", json=apt, headers=get_headers())
        if resp.status_code == 200:
            print(f"  ✓ Editado agendamento #{apt_id}")
            edited_count += 1
        else:
            errors.append(f"Appointment PUT {apt_id}: {resp.status_code}")
            print(f"  ✗ Erro ao editar agendamento #{apt_id}: {resp.status_code}")

    # Criar novo
    new_apt = {
        "patient_id": appointments[0]["patient_id"] if appointments else 5,
        "dentist_id": 4,
        "start_time": (datetime.now() + timedelta(days=7)).strftime('%Y-%m-%dT%H:%M:%SZ'),
        "end_time": (datetime.now() + timedelta(days=7, hours=1)).strftime('%Y-%m-%dT%H:%M:%SZ'),
        "type": "regular",
        "procedure": "Consulta de teste",
        "status": "scheduled",
        "notes": f"TESTE CRIADO - {datetime.now().strftime('%H:%M:%S')}"
    }

    resp = requests.post(f"{API_URL}/appointments", json=new_apt, headers=get_headers())
    if resp.status_code == 201:
        print(f"  ✓ Criado novo agendamento")
        created_count += 1
    else:
        errors.append(f"Appointment POST: {resp.status_code}")
        print(f"  ✗ Erro ao criar agendamento: {resp.status_code}")

def test_patients():
    global edited_count, created_count, errors
    print("\n=== TESTANDO: Pacientes ===")

    resp = requests.get(f"{API_URL}/patients?page=1&page_size=10", headers=get_headers())
    if resp.status_code != 200:
        errors.append(f"Patients GET: {resp.status_code}")
        return

    patients = resp.json().get("patients", [])
    if patients:
        # Editar
        patient = patients[0]
        patient_id = patient["id"]
        patient["notes"] = f"TESTE EDITADO - {datetime.now().strftime('%H:%M:%S')}"

        resp = requests.put(f"{API_URL}/patients/{patient_id}", json=patient, headers=get_headers())
        if resp.status_code == 200:
            print(f"  ✓ Editado paciente #{patient_id}")
            edited_count += 1
        else:
            errors.append(f"Patient PUT {patient_id}: {resp.status_code}")

    # Criar
    new_patient = {
        "name": f"Paciente Teste {datetime.now().strftime('%H%M%S')}",
        "email": f"teste{datetime.now().strftime('%H%M%S')}@teste.com",
        "phone": "(11) 99999-9999",
        "cpf": "000.000.000-00",
        "birth_date": "1990-01-01T00:00:00Z",
        "active": True
    }

    resp = requests.post(f"{API_URL}/patients", json=new_patient, headers=get_headers())
    if resp.status_code == 201:
        print(f"  ✓ Criado novo paciente")
        created_count += 1
    else:
        errors.append(f"Patient POST: {resp.status_code}")

def test_budgets():
    global edited_count, created_count, errors
    print("\n=== TESTANDO: Orçamentos ===")

    resp = requests.get(f"{API_URL}/budgets?page=1&page_size=10", headers=get_headers())
    if resp.status_code != 200:
        errors.append(f"Budgets GET: {resp.status_code}")
        return

    budgets = resp.json().get("budgets", [])
    if budgets:
        # Editar
        budget = budgets[0]
        budget_id = budget["id"]
        budget["notes"] = f"TESTE EDITADO - {datetime.now().strftime('%H:%M:%S')}"

        resp = requests.put(f"{API_URL}/budgets/{budget_id}", json=budget, headers=get_headers())
        if resp.status_code == 200:
            print(f"  ✓ Editado orçamento #{budget_id}")
            edited_count += 1
        else:
            errors.append(f"Budget PUT {budget_id}: {resp.status_code}")

def test_products():
    global edited_count, created_count, errors
    print("\n=== TESTANDO: Produtos ===")

    resp = requests.get(f"{API_URL}/products?page=1&page_size=10", headers=get_headers())
    if resp.status_code != 200:
        errors.append(f"Products GET: {resp.status_code}")
        return

    products = resp.json().get("products", [])
    if products:
        # Editar
        product = products[0]
        product_id = product["id"]
        product["notes"] = f"TESTE EDITADO - {datetime.now().strftime('%H:%M:%S')}"

        resp = requests.put(f"{API_URL}/products/{product_id}", json=product, headers=get_headers())
        if resp.status_code == 200:
            print(f"  ✓ Editado produto #{product_id}")
            edited_count += 1
        else:
            errors.append(f"Product PUT {product_id}: {resp.status_code}")

    # Criar
    new_product = {
        "name": f"Produto Teste {datetime.now().strftime('%H%M%S')}",
        "code": f"TST{datetime.now().strftime('%H%M%S')}",
        "category": "material",
        "quantity": 100,
        "minimum_stock": 10,
        "unit_price": 50.00,
        "active": True
    }

    resp = requests.post(f"{API_URL}/products", json=new_product, headers=get_headers())
    if resp.status_code == 201:
        print(f"  ✓ Criado novo produto")
        created_count += 1
    else:
        errors.append(f"Product POST: {resp.status_code}")

def test_tasks():
    global edited_count, created_count, errors
    print("\n=== TESTANDO: Tarefas ===")

    resp = requests.get(f"{API_URL}/tasks?page=1&page_size=10", headers=get_headers())
    if resp.status_code != 200:
        errors.append(f"Tasks GET: {resp.status_code}")
        return

    tasks = resp.json().get("tasks", [])
    if tasks:
        # Editar
        task = tasks[0]
        task_id = task["id"]
        task["description"] = f"TESTE EDITADO - {datetime.now().strftime('%H:%M:%S')}"

        resp = requests.put(f"{API_URL}/tasks/{task_id}", json=task, headers=get_headers())
        if resp.status_code == 200:
            print(f"  ✓ Editada tarefa #{task_id}")
            edited_count += 1
        else:
            errors.append(f"Task PUT {task_id}: {resp.status_code}")

def main():
    print("=" * 60)
    print("TESTE COMPLETO DE TODAS AS ABAS DO SISTEMA")
    print("=" * 60)

    if not login():
        return

    test_patients()
    test_appointments()
    test_budgets()
    test_products()
    test_tasks()

    print("\n" + "=" * 60)
    print("RESUMO DOS TESTES")
    print("=" * 60)
    print(f"✓ Total Editado: {edited_count}")
    print(f"✓ Total Criado: {created_count}")
    print(f"✗ Total Erros: {len(errors)}")

    if errors:
        print("\nERROS ENCONTRADOS:")
        for error in errors:
            print(f"  - {error}")
    else:
        print("\n✓ TODOS OS TESTES PASSARAM SEM ERROS!")

if __name__ == "__main__":
    main()
