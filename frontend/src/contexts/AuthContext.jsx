import React, { createContext, useState, useContext, useEffect } from 'react';
import { authAPI } from '../services/api';

const AuthContext = createContext({});

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [tenant, setTenant] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    const storedUser = localStorage.getItem('user');
    const storedTenant = localStorage.getItem('tenant');

    if (token && storedUser) {
      setUser(JSON.parse(storedUser));
      if (storedTenant) {
        setTenant(JSON.parse(storedTenant));
      }
      // Verify token is still valid
      authAPI.getMe()
        .then(response => {
          setUser(response.data.user);
          setTenant(response.data.tenant);
        })
        .catch(() => {
          logout();
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (credentials) => {
    const response = await authAPI.login(credentials);
    const { token, user: userData, tenant: tenantData } = response.data;

    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
    localStorage.setItem('tenant', JSON.stringify(tenantData));

    setUser(userData);
    setTenant(tenantData);

    return response.data;
  };

  const register = async (data) => {
    const response = await authAPI.register(data);
    return response.data;
  };

  const createTenant = async (data) => {
    const response = await authAPI.createTenant(data);
    const { token, admin, tenant: tenantData } = response.data;

    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(admin));
    localStorage.setItem('tenant', JSON.stringify(tenantData));

    setUser(admin);
    setTenant(tenantData);

    return response.data;
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('tenant');
    setUser(null);
    setTenant(null);
  };

  const updateUser = (userData) => {
    setUser(userData);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  const value = {
    user,
    tenant,
    loading,
    login,
    register,
    createTenant,
    logout,
    updateUser,
    isAuthenticated: !!user,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
