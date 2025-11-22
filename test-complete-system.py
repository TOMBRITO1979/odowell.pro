#!/usr/bin/env python3
"""
Complete system test - Tests all modules by creating and editing data
"""
import requests
import json
import time
from datetime import datetime, timedelta

API_URL = "https://drapi.crwell.pro/api"

class SystemTester:
    def __init__(self):
        self.token = None
        self.tenant_id = None
        self.user_id = None
        self.stats = {
            'created': {},
            'edited': {},
            'total_modules': 0
        }

    def login(self, email, password):
        """Login and get JWT token"""
        print("🔐 Logging in...")
        response = requests.post(f"{API_URL}/auth/login", json={
            "email": email,
            "password": password
        })

        if response.status_code != 200:
            print(f"❌ Login failed: {response.text}")
            return False

        data = response.json()
        self.token = data['token']
        self.tenant_id = data['user']['tenant_id']
        self.user_id = data['user']['id']
        print(f"✅ Logged in as {data['user']['name']} (ID: {self.user_id}, Tenant: {self.tenant_id})")
        return True

    def headers(self):
        """Get headers with auth token"""
        return {
            "Authorization": f"Bearer {self.token}",
            "Content-Type": "application/json"
        }

    def test_patients(self):
        """Test Patients module"""
        print("\n📋 Testing PATIENTS module...")
        module = "patients"

        # Create patient
        patient_data = {
            "name": f"Test Patient {datetime.now().strftime('%H%M%S')}",
            "email": f"patient{int(time.time())}@test.com",
            "phone": "(11) 98765-4321",
            "cpf": "123.456.789-00",
            "birth_date": "1990-05-15T00:00:00Z",
            "address": "Rua Teste, 123",
            "city": "São Paulo",
            "state": "SP",
            "zip_code": "01234-567"
        }

        response = requests.post(f"{API_URL}/patients", json=patient_data, headers=self.headers())
        if response.status_code == 201:
            patient = response.json()['patient']
            patient_id = patient['id']
            print(f"  ✅ Created patient ID {patient_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit patient
            edit_data = {
                **patient_data,
                "phone": "(11) 91111-2222",
                "notes": "Patient updated via automated test"
            }
            response = requests.put(f"{API_URL}/patients/{patient_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited patient ID {patient_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return patient_id
        else:
            print(f"  ❌ Failed to create patient: {response.text}")

        return None

    def test_appointments(self, patient_id):
        """Test Appointments module"""
        print("\n📅 Testing APPOINTMENTS module...")
        module = "appointments"

        if not patient_id:
            print("  ⚠️  Skipping (no patient ID)")
            return None

        # Create appointment
        start_time = (datetime.now() + timedelta(days=3)).replace(hour=14, minute=0, second=0, microsecond=0)
        end_time = start_time + timedelta(hours=1)

        appointment_data = {
            "patient_id": patient_id,
            "title": "Consulta de Rotina",
            "start_time": start_time.strftime('%Y-%m-%dT%H:%M:%SZ'),
            "end_time": end_time.strftime('%Y-%m-%dT%H:%M:%SZ'),
            "status": "scheduled",
            "notes": "Test appointment"
        }

        response = requests.post(f"{API_URL}/appointments", json=appointment_data, headers=self.headers())
        if response.status_code == 201:
            appointment = response.json()['appointment']
            appointment_id = appointment['id']
            print(f"  ✅ Created appointment ID {appointment_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit appointment
            edit_data = {
                **appointment_data,
                "status": "confirmed",
                "notes": "Appointment confirmed via automated test"
            }
            response = requests.put(f"{API_URL}/appointments/{appointment_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited appointment ID {appointment_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return appointment_id
        else:
            print(f"  ❌ Failed to create appointment: {response.text}")

        return None

    def test_tasks(self):
        """Test Tasks module"""
        print("\n✅ Testing TASKS module...")
        module = "tasks"

        # Create task
        task_data = {
            "title": f"Test Task {datetime.now().strftime('%H:%M:%S')}",
            "description": "This is a test task created by automated testing",
            "due_date": (datetime.now() + timedelta(days=5)).date().isoformat(),
            "priority": "high",
            "status": "pending",
            "assigned_to": [self.user_id]
        }

        response = requests.post(f"{API_URL}/tasks", json=task_data, headers=self.headers())
        if response.status_code == 201:
            task = response.json()['task']
            task_id = task['id']
            print(f"  ✅ Created task ID {task_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit task
            edit_data = {
                **task_data,
                "status": "in_progress",
                "description": "Task updated - now in progress"
            }
            response = requests.put(f"{API_URL}/tasks/{task_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited task ID {task_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return task_id
        else:
            print(f"  ❌ Failed to create task: {response.text}")

        return None

    def test_medical_records(self, patient_id):
        """Test Medical Records module"""
        print("\n🏥 Testing MEDICAL RECORDS module...")
        module = "medical_records"

        if not patient_id:
            print("  ⚠️  Skipping (no patient ID)")
            return None

        # Create medical record
        record_data = {
            "patient_id": patient_id,
            "diagnosis": "Cárie no dente 16",
            "treatment": "Restauração com resina composta",
            "prescription": "Analgésico 500mg - 1cp 8/8h por 3 dias",
            "notes": "Patient reported sensitivity"
        }

        response = requests.post(f"{API_URL}/medical-records", json=record_data, headers=self.headers())
        if response.status_code == 201:
            result = response.json()
            record = result.get('medical_record') or result.get('record') or result
            record_id = record['id']
            print(f"  ✅ Created medical record ID {record_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit medical record
            edit_data = {
                **record_data,
                "notes": "Updated: Patient improved after treatment"
            }
            response = requests.put(f"{API_URL}/medical-records/{record_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited medical record ID {record_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return record_id
        else:
            print(f"  ❌ Failed to create medical record: {response.text}")

        return None

    def test_budgets(self, patient_id):
        """Test Budgets module"""
        print("\n💰 Testing BUDGETS module...")
        module = "budgets"

        if not patient_id:
            print("  ⚠️  Skipping (no patient ID)")
            return None

        # Create budget
        budget_data = {
            "patient_id": patient_id,
            "description": "Tratamento completo",
            "items": [
                {
                    "description": "Restauração",
                    "quantity": 2,
                    "unit_price": 250.00
                },
                {
                    "description": "Limpeza",
                    "quantity": 1,
                    "unit_price": 150.00
                }
            ],
            "discount": 50.00,
            "notes": "Orçamento teste"
        }

        response = requests.post(f"{API_URL}/budgets", json=budget_data, headers=self.headers())
        if response.status_code == 201:
            budget = response.json()['budget']
            budget_id = budget['id']
            print(f"  ✅ Created budget ID {budget_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit budget
            edit_data = {
                **budget_data,
                "discount": 100.00,
                "notes": "Budget updated with higher discount"
            }
            response = requests.put(f"{API_URL}/budgets/{budget_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited budget ID {budget_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return budget_id
        else:
            print(f"  ❌ Failed to create budget: {response.text}")

        return None

    def test_payments(self):
        """Test Payments module"""
        print("\n💳 Testing PAYMENTS module...")
        module = "payments"

        # Create payment
        payment_data = {
            "description": f"Payment Test {datetime.now().strftime('%H:%M:%S')}",
            "amount": 500.00,
            "payment_method": "credit_card",
            "payment_date": datetime.now().date().isoformat(),
            "status": "completed",
            "notes": "Test payment"
        }

        response = requests.post(f"{API_URL}/payments", json=payment_data, headers=self.headers())
        if response.status_code == 201:
            payment = response.json()['payment']
            payment_id = payment['id']
            print(f"  ✅ Created payment ID {payment_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit payment
            edit_data = {
                **payment_data,
                "notes": "Payment updated via automated test"
            }
            response = requests.put(f"{API_URL}/payments/{payment_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited payment ID {payment_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return payment_id
        else:
            print(f"  ❌ Failed to create payment: {response.text}")

        return None

    def test_products(self):
        """Test Products module"""
        print("\n📦 Testing PRODUCTS module...")
        module = "products"

        # Create product
        product_data = {
            "name": f"Product Test {datetime.now().strftime('%H%M%S')}",
            "sku": f"SKU{int(time.time())}",
            "description": "Test product",
            "unit_price": 50.00,
            "stock_quantity": 100,
            "min_stock": 10,
            "category": "Materials",
            "active": True
        }

        response = requests.post(f"{API_URL}/products", json=product_data, headers=self.headers())
        if response.status_code == 201:
            product = response.json()['product']
            product_id = product['id']
            print(f"  ✅ Created product ID {product_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit product
            edit_data = {
                **product_data,
                "unit_price": 55.00,
                "description": "Product updated - new price"
            }
            response = requests.put(f"{API_URL}/products/{product_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited product ID {product_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return product_id
        else:
            print(f"  ❌ Failed to create product: {response.text}")

        return None

    def test_suppliers(self):
        """Test Suppliers module"""
        print("\n🏢 Testing SUPPLIERS module...")
        module = "suppliers"

        # Create supplier
        supplier_data = {
            "name": f"Supplier Test {datetime.now().strftime('%H%M%S')}",
            "cnpj": "12.345.678/0001-90",
            "email": f"supplier{int(time.time())}@test.com",
            "phone": "(11) 3333-4444",
            "address": "Av. Test, 456",
            "city": "São Paulo",
            "state": "SP",
            "zip_code": "01234-567",
            "active": True
        }

        response = requests.post(f"{API_URL}/suppliers", json=supplier_data, headers=self.headers())
        if response.status_code == 201:
            supplier = response.json()['supplier']
            supplier_id = supplier['id']
            print(f"  ✅ Created supplier ID {supplier_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit supplier
            edit_data = {
                **supplier_data,
                "phone": "(11) 3333-5555",
                "notes": "Supplier updated via automated test"
            }
            response = requests.put(f"{API_URL}/suppliers/{supplier_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited supplier ID {supplier_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return supplier_id
        else:
            print(f"  ❌ Failed to create supplier: {response.text}")

        return None

    def test_stock_movements(self, product_id):
        """Test Stock Movements module"""
        print("\n📊 Testing STOCK MOVEMENTS module...")
        module = "stock_movements"

        if not product_id:
            print("  ⚠️  Skipping (no product ID)")
            return None

        # Create stock movement
        movement_data = {
            "product_id": product_id,
            "type": "entry",
            "quantity": 50,
            "reason": "purchase",
            "notes": "Test stock movement"
        }

        response = requests.post(f"{API_URL}/stock-movements", json=movement_data, headers=self.headers())
        if response.status_code == 201:
            result = response.json()
            # Try different possible keys
            movement_id = None
            if 'stock_movement' in result:
                movement_id = result['stock_movement']['id']
            elif 'movement' in result:
                movement_id = result['movement']['id']
            elif 'id' in result:
                movement_id = result['id']

            if movement_id:
                print(f"  ✅ Created stock movement ID {movement_id}")
                self.stats['created'][module] = self.stats['created'].get(module, 0) + 1
                return movement_id
        else:
            print(f"  ❌ Failed to create stock movement: {response.text}")

        return None

    def test_campaigns(self):
        """Test Campaigns module"""
        print("\n📢 Testing CAMPAIGNS module...")
        module = "campaigns"

        # Create campaign
        campaign_data = {
            "name": f"Campaign Test {datetime.now().strftime('%H:%M:%S')}",
            "message": "Hello! This is a test campaign message.",
            "scheduled_date": (datetime.now() + timedelta(days=2)).isoformat(),
            "status": "draft"
        }

        response = requests.post(f"{API_URL}/campaigns", json=campaign_data, headers=self.headers())
        if response.status_code == 201:
            campaign = response.json()['campaign']
            campaign_id = campaign['id']
            print(f"  ✅ Created campaign ID {campaign_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit campaign
            edit_data = {
                **campaign_data,
                "message": "Updated campaign message!",
                "status": "scheduled"
            }
            response = requests.put(f"{API_URL}/campaigns/{campaign_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited campaign ID {campaign_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return campaign_id
        else:
            print(f"  ❌ Failed to create campaign: {response.text}")

        return None

    def test_prescriptions(self, patient_id):
        """Test Prescriptions module"""
        print("\n💊 Testing PRESCRIPTIONS module...")
        module = "prescriptions"

        if not patient_id:
            print("  ⚠️  Skipping (no patient ID)")
            return None

        # Create prescription
        prescription_data = {
            "patient_id": patient_id,
            "medications": [
                {
                    "name": "Amoxicilina 500mg",
                    "dosage": "1 cápsula",
                    "frequency": "8/8h",
                    "duration": "7 dias"
                }
            ],
            "notes": "Test prescription"
        }

        response = requests.post(f"{API_URL}/prescriptions", json=prescription_data, headers=self.headers())
        if response.status_code == 201:
            prescription = response.json()['prescription']
            prescription_id = prescription['id']
            print(f"  ✅ Created prescription ID {prescription_id}")
            self.stats['created'][module] = self.stats['created'].get(module, 0) + 1

            # Edit prescription
            edit_data = {
                **prescription_data,
                "notes": "Prescription updated - patient allergies checked"
            }
            response = requests.put(f"{API_URL}/prescriptions/{prescription_id}", json=edit_data, headers=self.headers())
            if response.status_code == 200:
                print(f"  ✅ Edited prescription ID {prescription_id}")
                self.stats['edited'][module] = self.stats['edited'].get(module, 0) + 1
                return prescription_id
        else:
            print(f"  ❌ Failed to create prescription: {response.text}")

        return None

    def run_tests(self, email, password):
        """Run all tests"""
        print("=" * 60)
        print("🧪 STARTING COMPLETE SYSTEM TEST")
        print("=" * 60)

        if not self.login(email, password):
            return

        # Test all modules
        patient_id = self.test_patients()
        appointment_id = self.test_appointments(patient_id)
        task_id = self.test_tasks()
        record_id = self.test_medical_records(patient_id)
        budget_id = self.test_budgets(patient_id)
        payment_id = self.test_payments()
        product_id = self.test_products()
        supplier_id = self.test_suppliers()
        movement_id = self.test_stock_movements(product_id)
        campaign_id = self.test_campaigns()
        prescription_id = self.test_prescriptions(patient_id)

        # Count modules tested
        self.stats['total_modules'] = len(self.stats['created'])

        # Print summary
        print("\n" + "=" * 60)
        print("📊 TEST SUMMARY")
        print("=" * 60)

        total_created = sum(self.stats['created'].values())
        total_edited = sum(self.stats['edited'].values())

        print(f"\n✅ Total records CREATED: {total_created}")
        for module, count in self.stats['created'].items():
            print(f"   - {module}: {count}")

        print(f"\n✏️  Total records EDITED: {total_edited}")
        for module, count in self.stats['edited'].items():
            print(f"   - {module}: {count}")

        print(f"\n📁 Total modules tested: {self.stats['total_modules']}")
        print("\n" + "=" * 60)

if __name__ == "__main__":
    tester = SystemTester()

    # Use admin credentials - you'll need to provide these
    EMAIL = input("Enter admin email: ")
    PASSWORD = input("Enter admin password: ")

    tester.run_tests(EMAIL, PASSWORD)
