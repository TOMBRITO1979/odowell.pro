import React, { createContext, useState, useContext, useEffect } from 'react';
import { jwtDecode } from 'jwt-decode';
import { authAPI } from '../services/api';

const AuthContext = createContext({});

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [tenant, setTenant] = useState(null);
  const [permissions, setPermissions] = useState({});
  const [loading, setLoading] = useState(true);

  // Helper function to extract permissions from JWT
  const extractPermissions = (token) => {
    try {
      const decoded = jwtDecode(token);
      return decoded.permissions || {};
    } catch (error) {
      return {};
    }
  };

  useEffect(() => {
    let mounted = true;

    const initializeAuth = async () => {
      const token = localStorage.getItem('token');
      const storedUser = localStorage.getItem('user');
      const storedTenant = localStorage.getItem('tenant');

      if (token && storedUser) {
        if (mounted) {
          setUser(JSON.parse(storedUser));
          if (storedTenant) {
            setTenant(JSON.parse(storedTenant));
          }
          // Extract permissions from JWT
          const perms = extractPermissions(token);
          setPermissions(perms);
        }

        // Verify token is still valid
        try {
          const response = await authAPI.getMe();
          if (mounted) {
            setUser(response.data.user);
            setTenant(response.data.tenant);
          }
        } catch {
          if (mounted) {
            logout();
          }
        } finally {
          if (mounted) {
            setLoading(false);
          }
        }
      } else {
        if (mounted) {
          setLoading(false);
        }
      }
    };

    initializeAuth();

    return () => {
      mounted = false;
    };
  }, []);

  const login = async (credentials) => {
    const response = await authAPI.login(credentials);
    const { token, user: userData, tenant: tenantData } = response.data;

    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(userData));
    localStorage.setItem('tenant', JSON.stringify(tenantData));

    // Extract and set permissions
    const perms = extractPermissions(token);
    setPermissions(perms);

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
    // NOTE: No longer auto-logging in - user must verify email first
    return response.data;
  };

  const logout = async () => {
    // Clear httpOnly cookie via API
    try {
      await authAPI.logout();
    } catch (error) {
      // Ignore errors - we're logging out anyway
    }
    // Clear local storage
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('tenant');
    setUser(null);
    setTenant(null);
    setPermissions({});
  };

  const updateUser = (userData) => {
    setUser(userData);
    localStorage.setItem('user', JSON.stringify(userData));
  };

  const updateTenant = (tenantData) => {
    setTenant(tenantData);
    localStorage.setItem('tenant', JSON.stringify(tenantData));
  };

  const value = {
    user,
    tenant,
    permissions,
    loading,
    login,
    register,
    createTenant,
    logout,
    updateUser,
    updateTenant,
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

// Hook to check user permissions
export const usePermission = () => {
  const { user, permissions } = useAuth();

  // Check if user has a specific permission
  const hasPermission = (module, action) => {
    // Admins have all permissions
    if (user?.role === 'admin') {
      return true;
    }

    // Check if user has the specific permission
    return permissions?.[module]?.[action] === true;
  };

  // Check if user has at least one permission for a module
  const hasAnyPermission = (module) => {
    if (user?.role === 'admin') {
      return true;
    }

    const modulePerms = permissions?.[module];
    if (!modulePerms) return false;

    return Object.values(modulePerms).some(perm => perm === true);
  };

  // Check if user can view a module (has view permission or any other permission)
  const canView = (module) => {
    return hasPermission(module, 'view') || hasAnyPermission(module);
  };

  // Check if user can create in a module
  const canCreate = (module) => {
    return hasPermission(module, 'create');
  };

  // Check if user can edit in a module
  const canEdit = (module) => {
    return hasPermission(module, 'edit');
  };

  // Check if user can delete in a module
  const canDelete = (module) => {
    return hasPermission(module, 'delete');
  };

  return {
    hasPermission,
    hasAnyPermission,
    canView,
    canCreate,
    canEdit,
    canDelete,
    permissions,
    isAdmin: user?.role === 'admin',
    isSuperAdmin: user?.is_super_admin === true,
  };
};
