import React, { useState, useEffect } from 'react';
import { Card, Radio, Button, Space, Tag, Tooltip, message } from 'antd';
import {
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  DeleteOutlined,
  ToolOutlined
} from '@ant-design/icons';
import './Odontogram.css';

const Odontogram = ({ value, onChange, readOnly = false }) => {
  const [odontogram, setOdontogram] = useState({});
  const [selectedTooth, setSelectedTooth] = useState(null);

  // Initialize odontogram from value prop
  useEffect(() => {
    if (value) {
      try {
        const parsed = typeof value === 'string' ? JSON.parse(value) : value;
        setOdontogram(parsed || {});
      } catch (e) {
        console.error('Error parsing odontogram:', e);
        setOdontogram({});
      }
    }
  }, [value]);

  // Notify parent of changes
  useEffect(() => {
    if (onChange && Object.keys(odontogram).length > 0) {
      onChange(JSON.stringify(odontogram));
    }
  }, [odontogram, onChange]);

  // Tooth numbering (FDI notation)
  const upperRight = [18, 17, 16, 15, 14, 13, 12, 11];
  const upperLeft = [21, 22, 23, 24, 25, 26, 27, 28];
  const lowerRight = [48, 47, 46, 45, 44, 43, 42, 41];
  const lowerLeft = [31, 32, 33, 34, 35, 36, 37, 38];

  const toothStatuses = [
    { value: 'healthy', label: 'Saudável', color: '#52c41a', icon: <CheckCircleOutlined /> },
    { value: 'cavity', label: 'Cárie', color: '#faad14', icon: <WarningOutlined /> },
    { value: 'restoration', label: 'Restauração', color: '#1890ff', icon: <ToolOutlined /> },
    { value: 'missing', label: 'Ausente', color: '#f5222d', icon: <CloseCircleOutlined /> },
    { value: 'root_canal', label: 'Canal', color: '#722ed1', icon: <ToolOutlined /> },
    { value: 'crown', label: 'Coroa', color: '#13c2c2', icon: <ToolOutlined /> },
    { value: 'implant', label: 'Implante', color: '#2f54eb', icon: <ToolOutlined /> },
  ];

  const getToothStatus = (toothNumber) => {
    const tooth = odontogram[toothNumber];
    if (!tooth || !tooth.status) return null;
    return toothStatuses.find(s => s.value === tooth.status);
  };

  const handleToothClick = (toothNumber) => {
    if (readOnly) return;
    setSelectedTooth(toothNumber);
  };

  const handleStatusChange = (status) => {
    if (!selectedTooth) return;

    const newOdontogram = { ...odontogram };
    if (!newOdontogram[selectedTooth]) {
      newOdontogram[selectedTooth] = { status: '', procedures: [] };
    }
    newOdontogram[selectedTooth].status = status;
    setOdontogram(newOdontogram);
    message.success(`Dente ${selectedTooth} marcado como ${toothStatuses.find(s => s.value === status)?.label}`);
  };

  const handleClearTooth = () => {
    if (!selectedTooth) return;

    const newOdontogram = { ...odontogram };
    delete newOdontogram[selectedTooth];
    setOdontogram(newOdontogram);
    setSelectedTooth(null);
    message.success(`Dente ${selectedTooth} limpo`);
  };

  const renderTooth = (toothNumber) => {
    const status = getToothStatus(toothNumber);
    const isSelected = selectedTooth === toothNumber;

    return (
      <Tooltip
        key={toothNumber}
        title={status ? `${toothNumber} - ${status.label}` : `Dente ${toothNumber}`}
      >
        <div
          className={`tooth ${isSelected ? 'tooth-selected' : ''} ${readOnly ? 'tooth-readonly' : ''}`}
          style={{
            backgroundColor: status ? status.color : '#fff',
            border: `2px solid ${isSelected ? '#1890ff' : (status ? status.color : '#d9d9d9')}`,
            color: status ? '#fff' : '#000',
            cursor: readOnly ? 'default' : 'pointer',
          }}
          onClick={() => handleToothClick(toothNumber)}
        >
          <div className="tooth-number">{toothNumber}</div>
          {status && <div className="tooth-icon">{status.icon}</div>}
        </div>
      </Tooltip>
    );
  };

  const content = (
    <>
      <div className="odontogram-grid">
        {/* Upper teeth */}
        <div className="teeth-row upper">
          <div className="teeth-quadrant upper-right">
            {upperRight.map(renderTooth)}
          </div>
          <div className="teeth-divider"></div>
          <div className="teeth-quadrant upper-left">
            {upperLeft.map(renderTooth)}
          </div>
        </div>

        {/* Lower teeth */}
        <div className="teeth-row lower">
          <div className="teeth-quadrant lower-right">
            {lowerRight.map(renderTooth)}
          </div>
          <div className="teeth-divider"></div>
          <div className="teeth-quadrant lower-left">
            {lowerLeft.map(renderTooth)}
          </div>
        </div>
      </div>

      {!readOnly && selectedTooth && (
        <Card
          size="small"
          title={`Dente ${selectedTooth} selecionado`}
          className="tooth-selector"
          extra={
            <Button
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={handleClearTooth}
            >
              Limpar
            </Button>
          }
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <div>Selecione o status:</div>
            <Radio.Group
              value={odontogram[selectedTooth]?.status}
              onChange={(e) => handleStatusChange(e.target.value)}
              style={{ width: '100%' }}
            >
              <Space direction="vertical" style={{ width: '100%' }}>
                {toothStatuses.map(status => (
                  <Radio key={status.value} value={status.value}>
                    <Space>
                      <span style={{ color: status.color }}>{status.icon}</span>
                      <span>{status.label}</span>
                    </Space>
                  </Radio>
                ))}
              </Space>
            </Radio.Group>
          </Space>
        </Card>
      )}

      {!readOnly && !selectedTooth && (
        <div style={{ textAlign: 'center', padding: '20px', color: '#999' }}>
          Clique em um dente para marcar seu status
        </div>
      )}

      {/* Legend */}
      <div className="odontogram-legend">
        <strong>Legenda:</strong>
        <Space wrap style={{ marginTop: 8 }}>
          {toothStatuses.map(status => (
            <Tag key={status.value} color={status.color}>
              {status.icon} {status.label}
            </Tag>
          ))}
        </Space>
      </div>
    </>
  );

  if (readOnly) {
    return (
      <div className="odontogram-container odontogram-readonly">
        {content}
      </div>
    );
  }

  return (
    <div className="odontogram-container">
      <Card title="Odontograma" className="odontogram-card">
        {content}
      </Card>
    </div>
  );
};

export default Odontogram;
