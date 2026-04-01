import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Button, Icon, Menu } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import {
  ADMIN_MENU_GROUPS,
  isAdminGroupActive,
  isAdminRouteActive,
} from '../constants/adminMenu';

const SIDEBAR_GROUP_COLLAPSED_STORAGE_KEY =
  'router_admin_sidebar_group_collapsed_v1';

const buildDefaultCollapsedState = () => {
  const defaults = {};
  ADMIN_MENU_GROUPS.forEach((group) => {
    defaults[group.key] = false;
  });
  return defaults;
};

const buildInitialCollapsedState = () => {
  const defaults = buildDefaultCollapsedState();
  if (typeof window === 'undefined') {
    return defaults;
  }
  const raw = (localStorage.getItem(SIDEBAR_GROUP_COLLAPSED_STORAGE_KEY) || '')
    .trim();
  if (raw === '') {
    return defaults;
  }
  try {
    const parsed = JSON.parse(raw);
    if (!parsed || typeof parsed !== 'object') {
      return defaults;
    }
    return Object.keys(defaults).reduce((result, key) => {
      result[key] = Boolean(parsed[key]);
      return result;
    }, {});
  } catch (error) {
    return defaults;
  }
};

const AdminSidebar = ({ compact = false, onToggleCompact }) => {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const [collapsedGroups, setCollapsedGroups] = useState(
    buildInitialCollapsedState,
  );

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    localStorage.setItem(
      SIDEBAR_GROUP_COLLAPSED_STORAGE_KEY,
      JSON.stringify(collapsedGroups),
    );
  }, [collapsedGroups]);

  const toggleGroup = (groupKey) => {
    setCollapsedGroups((prev) => ({
      ...prev,
      [groupKey]: !prev[groupKey],
    }));
  };

  const isGroupCollapsed = (group) => {
    const groupActive = isAdminGroupActive(location, group);
    if (groupActive) {
      return false;
    }
    return Boolean(collapsedGroups[group.key]);
  };

  return (
    <Menu vertical fluid className='router-admin-sidebar-menu'>
      <Menu.Item className='router-admin-sidebar-toolbar'>
        <Button
          basic
          icon
          size='mini'
          type='button'
          className='router-admin-sidebar-toggle'
          title={
            compact
              ? t('header.sidebar_expand')
              : t('header.sidebar_compact')
          }
          onClick={() => {
            if (typeof onToggleCompact === 'function') {
              onToggleCompact();
            }
          }}
        >
          <Icon name={compact ? 'angle double right' : 'angle double left'} />
        </Button>
      </Menu.Item>
      {ADMIN_MENU_GROUPS.map((group) => {
        const groupActive = isAdminGroupActive(location, group);
        const collapsed = isGroupCollapsed(group);
        return (
          <Menu.Item
            key={group.key}
            className={`router-admin-sidebar-group ${groupActive ? 'active' : ''}`}
          >
            <div
              className='router-admin-sidebar-group-header'
              role='button'
              tabIndex={0}
              onClick={() => toggleGroup(group.key)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ' ') {
                  event.preventDefault();
                  toggleGroup(group.key);
                }
              }}
            >
              <span
                className='router-admin-sidebar-group-title'
                title={t(group.name)}
              >
                <Icon name={group.icon} />
                <span className='router-admin-sidebar-group-label'>
                  {t(group.name)}
                </span>
              </span>
              <Icon name={collapsed ? 'angle right' : 'angle down'} />
            </div>
            {!collapsed ? (
              <Menu.Menu>
                {group.items.map((item) => {
                  const active = isAdminRouteActive(location, item.to);
                  return (
                    <Menu.Item
                      key={item.to}
                      active={active}
                      onClick={() => navigate(item.to)}
                      className='router-admin-sidebar-item'
                      title={t(item.name)}
                    >
                      <Icon name={item.icon} />
                      <span className='router-admin-sidebar-item-label'>
                        {t(item.name)}
                      </span>
                    </Menu.Item>
                  );
                })}
              </Menu.Menu>
            ) : null}
          </Menu.Item>
        );
      })}
    </Menu>
  );
};

export default AdminSidebar;
