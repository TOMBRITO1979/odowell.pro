import React, { useState, useEffect } from 'react';
import { Drawer, Checkbox, Button, message, Spin, Card, Row, Col, Space } from 'antd';
import { permissionsAPI } from '../../services/api';

const UserPermissions = ({ user, visible, onClose }) => {
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [modules, setModules] = useState([]);
  const [permissions, setPermissions] = useState({});

  useEffect(() => {
    if (visible && user) {
      fetchData();
    }
  }, [visible, user]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [modulesRes, permsRes] = await Promise.all([
        permissionsAPI.getModules(),
        permissionsAPI.getUserPermissions(user.id),
      ]);
      setModules(modulesRes.data.modules || []);
      setPermissions(permsRes.data.permissions || {});
    } catch (error) {
      message.error('Erro ao carregar permissões');
    } finally {
      setLoading(false);
    }
  };

  const handlePermissionChange = (moduleCode, action, checked) => {
    setPermissions(prev => ({
      ...prev,
      [moduleCode]: {
        ...(prev[moduleCode] || {}),
        [action]: checked,
      },
    }));
  };

  const handleApplyDefaults = async () => {
    try {
      const response = await permissionsAPI.getDefaultRolePermissions(user.role);
      setPermissions(response.data.permissions || {});
      message.success('Permissões padrão aplicadas');
    } catch (error) {
      message.error('Erro ao aplicar permissões padrão');
    }
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      await permissionsAPI.updateUserPermissions(user.id, permissions);
      message.success('Permissões atualizadas com sucesso');
      onClose();
    } catch (error) {
      message.error('Erro ao salvar permissões');
    } finally {
      setSaving(false);
    }
  };

  const actions = ['view', 'create', 'edit', 'delete'];
  const actionLabels = {
    view: 'Visualizar',
    create: 'Criar',
    edit: 'Editar',
    delete: 'Deletar',
  };

  return (
    <Drawer
      title={`Permissões de ${user?.name}`}
      open={visible}
      onClose={onClose}
      width={720}
      footer={
        <Space>
          <Button onClick={onClose}>Cancelar</Button>
          <Button onClick={handleApplyDefaults}>
            Aplicar Permissões Padrão ({user?.role})
          </Button>
          <Button type="primary" onClick={handleSave} loading={saving}>
            Salvar
          </Button>
        </Space>
      }
    >
      {loading ? (
        <Spin />
      ) : (
        <div>
          {modules.map(module => (
            <Card key={module.code} title={module.name} size="small" style={{ marginBottom: 16 }}>
              <Row gutter={16}>
                {actions.map(action => (
                  <Col span={6} key={action}>
                    <Checkbox
                      checked={permissions[module.code]?.[action] === true}
                      onChange={(e) => handlePermissionChange(module.code, action, e.target.checked)}
                    >
                      {actionLabels[action]}
                    </Checkbox>
                  </Col>
                ))}
              </Row>
            </Card>
          ))}
        </div>
      )}
    </Drawer>
  );
};

export default UserPermissions;
