# Mapeamento de Rotas e Permissões

## Módulo: patients (Pacientes)
- POST   /patients         -> patients:create
- GET    /patients         -> patients:view
- GET    /patients/:id     -> patients:view
- PUT    /patients/:id     -> patients:edit
- DELETE /patients/:id     -> patients:delete

## Módulo: appointments (Agendamentos)
- POST   /appointments            -> appointments:create
- GET    /appointments            -> appointments:view
- GET    /appointments/:id        -> appointments:view
- PUT    /appointments/:id        -> appointments:edit
- DELETE /appointments/:id        -> appointments:delete
- PATCH  /appointments/:id/status -> appointments:edit

## Módulo: medical_records (Prontuários)
- POST   /medical-records         -> medical_records:create
- GET    /medical-records         -> medical_records:view
- GET    /medical-records/:id     -> medical_records:view
- PUT    /medical-records/:id     -> medical_records:edit
- DELETE /medical-records/:id     -> medical_records:delete
- GET    /medical-records/:id/pdf -> medical_records:view

## Módulo: prescriptions (Receituário)
- POST   /prescriptions            -> prescriptions:create
- GET    /prescriptions            -> prescriptions:view
- GET    /prescriptions/:id        -> prescriptions:view
- PUT    /prescriptions/:id        -> prescriptions:edit
- DELETE /prescriptions/:id        -> prescriptions:delete
- POST   /prescriptions/:id/issue  -> prescriptions:edit
- POST   /prescriptions/:id/print  -> prescriptions:view
- GET    /prescriptions/:id/pdf    -> prescriptions:view

## Módulo: exams (Exames)
- POST   /exams               -> exams:create
- GET    /exams               -> exams:view
- GET    /exams/:id           -> exams:view
- PUT    /exams/:id           -> exams:edit
- DELETE /exams/:id           -> exams:delete
- GET    /exams/:id/download  -> exams:view

## Módulo: budgets (Orçamentos)
- POST   /budgets                              -> budgets:create
- GET    /budgets                              -> budgets:view
- GET    /budgets/:id                          -> budgets:view
- PUT    /budgets/:id                          -> budgets:edit
- DELETE /budgets/:id                          -> budgets:delete
- GET    /budgets/:id/pdf                      -> budgets:view
- GET    /budgets/:id/payment/:payment_id/receipt -> budgets:view

## Módulo: payments (Pagamentos)
- POST   /payments           -> payments:create
- GET    /payments           -> payments:view
- GET    /payments/:id       -> payments:view
- PUT    /payments/:id       -> payments:edit
- DELETE /payments/:id       -> payments:delete
- GET    /payments/pdf/export -> payments:view

## Módulo: products (Produtos)
- POST   /products           -> products:create
- GET    /products           -> products:view
- GET    /products/:id       -> products:view
- PUT    /products/:id       -> products:edit
- DELETE /products/:id       -> products:delete
- GET    /products/low-stock -> products:view

## Módulo: suppliers (Fornecedores)
- POST   /suppliers     -> suppliers:create
- GET    /suppliers     -> suppliers:view
- GET    /suppliers/:id -> suppliers:view
- PUT    /suppliers/:id -> suppliers:edit
- DELETE /suppliers/:id -> suppliers:delete

## Módulo: stock_movements (Movimentações)
- POST   /stock-movements -> stock_movements:create
- GET    /stock-movements -> stock_movements:view

## Módulo: campaigns (Campanhas)
- POST   /campaigns         -> campaigns:create
- GET    /campaigns         -> campaigns:view
- GET    /campaigns/:id     -> campaigns:view
- PUT    /campaigns/:id     -> campaigns:edit
- DELETE /campaigns/:id     -> campaigns:delete
- POST   /campaigns/:id/send -> campaigns:edit

## Módulo: reports (Relatórios)
- GET /reports/dashboard         -> reports:view
- GET /reports/revenue           -> reports:view
- GET /reports/procedures        -> reports:view
- GET /reports/attendance        -> reports:view
- GET /reports/revenue/pdf       -> reports:view
- GET /reports/attendance/pdf    -> reports:view
- GET /reports/procedures/pdf    -> reports:view
- GET /reports/revenue/excel     -> reports:view
- GET /reports/attendance/excel  -> reports:view
- GET /reports/procedures/excel  -> reports:view

## Módulo: settings (Configurações)
- GET /settings -> settings:view
- PUT /settings -> settings:edit

## Rotas que NÃO precisam de middleware de permissões:
- /api/tenants (público)
- /api/auth/login (público)
- /api/auth/me (só auth)
- /api/auth/profile (só auth)
- /api/auth/password (só auth)
- /api/users/* (admin only - já tem middleware separado)
- /api/permissions/* (admin only - já tem middleware separado)

