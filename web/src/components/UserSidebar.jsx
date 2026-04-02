import React, { useMemo } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Icon, Menu } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';

const isUserRouteActive = (location, to) => {
  const [path] = String(to || '').split('?');
  if (!path) {
    return false;
  }
  return location.pathname === path || location.pathname.startsWith(`${path}/`);
};

const buildUserMenuItems = (includeChat = false) => {
  const items = [
    {
      name: 'header.dashboard',
      to: '/workspace/dashboard',
      icon: 'chart bar',
    },
    {
      name: 'header.token',
      to: '/workspace/token',
      icon: 'key',
    },
  ];

  if (includeChat) {
    items.push({
      name: 'header.chat',
      to: '/workspace/chat',
      icon: 'comments',
    });
  }

  items.push(
    {
      name: 'header.topup',
      to: '/workspace/topup',
      icon: 'cart',
    },
    {
      name: 'header.log',
      to: '/workspace/log',
      icon: 'book',
    },
    {
      name: 'header.task',
      to: '/workspace/task',
      icon: 'tasks',
    },
    {
      name: 'header.setting',
      to: '/workspace/setting',
      icon: 'setting',
    },
  );

  return items;
};

const UserSidebar = ({ compact = false }) => {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const includeChat = Boolean(localStorage.getItem('chat_link'));
  const menuItems = useMemo(
    () => buildUserMenuItems(includeChat),
    [includeChat],
  );

  return (
    <Menu vertical fluid className='router-admin-sidebar-menu'>
      {menuItems.map((item) => {
        const active = isUserRouteActive(location, item.to);
        return (
          <Menu.Item
            key={item.to}
            active={active}
            onClick={() => navigate(item.to)}
            className={`router-admin-sidebar-group router-user-sidebar-item ${active ? 'active' : ''}`}
            title={t(item.name)}
          >
            <span className='router-admin-sidebar-item-content'>
              <Icon
                name={item.icon}
                className='router-admin-sidebar-item-icon'
              />
              {!compact ? (
                <span className='router-admin-sidebar-item-label'>
                  {t(item.name)}
                </span>
              ) : null}
            </span>
          </Menu.Item>
        );
      })}
    </Menu>
  );
};

export default UserSidebar;
