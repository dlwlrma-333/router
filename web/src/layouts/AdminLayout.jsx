import React, { useEffect, useMemo, useState } from 'react';
import { Outlet } from 'react-router-dom';
import { Container } from 'semantic-ui-react';
import Footer from '../components/Footer';
import Header from '../components/Header';
import AdminSidebar from '../components/AdminSidebar';

const SIDEBAR_COMPACT_STORAGE_KEY = 'router_admin_sidebar_compact_v1';

const AdminLayout = () => {
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
      <div className='router-admin-shell'>
        <aside
          className={`router-admin-sidebar ${sidebarCompact ? 'compact' : ''}`}
        >
          <AdminSidebar
            compact={sidebarCompact}
            onToggleCompact={() =>
              setSidebarCompact((previous) => !previous)
            }
          />
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
