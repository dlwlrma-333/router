import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Grid, Header, Menu } from 'semantic-ui-react';
import { useLocation, useSearchParams } from 'react-router-dom';
import SystemSetting from '../../components/SystemSetting';
import { isRoot } from '../../helpers';
import OtherSetting from '../../components/OtherSetting';
import PersonalSetting from '../../components/PersonalSetting';
import OperationSetting from '../../components/OperationSetting';

const Setting = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const isAdminWorkspace = location.pathname.startsWith('/admin/');

  if (!isAdminWorkspace) {
    return (
      <div className='dashboard-container'>
        <Card fluid className='chart-card'>
          <Card.Content>
            <Card.Header className='header router-page-title'>
              {t('setting.title')}
            </Card.Header>
            <PersonalSetting />
          </Card.Content>
        </Card>
      </div>
    );
  }

  const menuGroups = [];

  if (isRoot()) {
    menuGroups.push({
      key: 'operation',
      label: t('setting.tabs.operation'),
      sections: [
        { key: 'quota', label: t('setting.operation.quota.title') },
        { key: 'monitor', label: t('setting.operation.monitor.title') },
        { key: 'log', label: t('setting.operation.log.title') },
        { key: 'general', label: t('setting.operation.general.title') },
        { key: 'billing', label: t('setting.operation.billing.title') },
      ],
    });
    menuGroups.push({
      key: 'system',
      label: t('setting.tabs.system'),
      sections: [
        { key: 'general', label: t('setting.system.general.title') },
        { key: 'smtp', label: t('setting.system.smtp.title') },
        { key: 'login', label: t('setting.system.login.title') },
      ],
    });
    menuGroups.push({
      key: 'other',
      label: t('setting.tabs.other'),
      sections: [
        { key: 'notice', label: t('setting.system.notice', '站点公告') },
        { key: 'content', label: t('setting.other.content.title') },
      ],
    });
  }

  const tabKeys = menuGroups.map((item) => item.key);
  const requestedTab = (searchParams.get('tab') || '').trim().toLowerCase();
  const activeTab =
    tabKeys.includes(requestedTab) && requestedTab !== ''
      ? requestedTab
      : tabKeys[0] || '';
  const activeGroup = menuGroups.find((item) => item.key === activeTab);
  const sectionKeys = (activeGroup?.sections || []).map((item) => item.key);
  const requestedSection = (searchParams.get('section') || '')
    .trim()
    .toLowerCase();
  const activeSection =
    sectionKeys.includes(requestedSection) && requestedSection !== ''
      ? requestedSection
      : sectionKeys[0] || '';

  const goToSection = (tab, section) => {
    const nextParams = new URLSearchParams(searchParams);
    nextParams.set('tab', tab);
    nextParams.set('section', section);
    setSearchParams(nextParams);
  };

  const renderContent = () => {
    if (activeTab === 'operation') {
      return <OperationSetting section={activeSection} />;
    }
    if (activeTab === 'system') {
      return <SystemSetting section={activeSection} />;
    }
    if (activeTab === 'other') {
      return <OtherSetting section={activeSection} />;
    }
    return <div className='router-empty-cell'>{t('setting.empty_admin', '暂无可配置项')}</div>;
  };

  const activeSectionLabel =
    activeGroup?.sections?.find((item) => item.key === activeSection)?.label ||
    '';

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='header router-page-title'>
            {t('setting.title')}
          </Card.Header>
          {menuGroups.length > 0 ? (
            <Grid stackable columns={2} className='router-settings-layout'>
              <Grid.Column width={4}>
                <Menu fluid vertical className='router-settings-menu'>
                  {menuGroups.map((group) => (
                    <Menu.Item key={group.key} className='router-settings-menu-group'>
                      <Menu.Header>{group.label}</Menu.Header>
                      <Menu.Menu>
                        {group.sections.map((section) => (
                          <Menu.Item
                            key={`${group.key}-${section.key}`}
                            active={
                              activeTab === group.key &&
                              activeSection === section.key
                            }
                            onClick={() => goToSection(group.key, section.key)}
                          >
                            {section.label}
                          </Menu.Item>
                        ))}
                      </Menu.Menu>
                    </Menu.Item>
                  ))}
                </Menu>
              </Grid.Column>
              <Grid.Column width={12}>
                {activeSectionLabel ? (
                  <Header as='h3' className='router-section-title'>
                    {activeSectionLabel}
                  </Header>
                ) : null}
                {renderContent()}
              </Grid.Column>
            </Grid>
          ) : (
            <div className='router-empty-cell'>
              {t('setting.empty_admin', '暂无可配置项')}
            </div>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default Setting;
