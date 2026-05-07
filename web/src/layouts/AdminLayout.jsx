import React, { useEffect, useMemo, useState } from 'react';
import { Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button, Container, Icon } from 'semantic-ui-react';
import Footer from '../components/Footer';
import Header from '../components/Header';
import AdminSidebar from '../components/AdminSidebar';

const SIDEBAR_COMPACT_STORAGE_KEY = 'router_admin_sidebar_compact_v1';

const AdminLayout = () => {
  const { t } = useTranslation();
  const initialCompact = useMemo(() => {
    if (typeof window === 'undefined') {
      return false;
    }
    const raw = (localStorage.getItem(SIDEBAR_COMPACT_STORAGE_KEY) || '')
      .trim()
      .toLowerCase();
    return raw === '1' || raw === 'true';
  }, []);
  const [sidebarCompact, setSidebarCompact] = useState(initialCompact);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }
    localStorage.setItem(
      SIDEBAR_COMPACT_STORAGE_KEY,
      sidebarCompact ? '1' : '0',
    );
  }, [sidebarCompact]);

  return (
    <>
      <Header workspace='admin' hideNavButtons />
      <div className={`router-admin-shell ${sidebarCompact ? 'compact' : ''}`}>
        <aside
          className={`router-admin-sidebar ${sidebarCompact ? 'compact' : ''}`}
        >
          <div className='router-admin-sidebar-toolbar'>
            <Button
              type='button'
              basic
              icon
              size='small'
              className='router-admin-sidebar-toggle'
              title={
                sidebarCompact
                  ? t('header.sidebar_expand')
                  : t('header.sidebar_compact')
              }
              aria-label={
                sidebarCompact
                  ? t('header.sidebar_expand')
                  : t('header.sidebar_compact')
              }
              onClick={() => setSidebarCompact((previous) => !previous)}
            >
              <Icon
                name={
                  sidebarCompact ? 'angle double right' : 'angle double left'
                }
              />
            </Button>
          </div>
          <AdminSidebar compact={sidebarCompact} />
        </aside>
        <Container className='main-content router-admin-main'>
          <Outlet />
        </Container>
      </div>
      <Footer />
    </>
  );
};

export default AdminLayout;
