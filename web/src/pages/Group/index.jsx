import React from 'react';
import { Card } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import GroupsManager from '../../components/GroupsManager';

const Group = () => {
  const { t } = useTranslation();

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header
            className='header'
            style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
          >
            <span>{t('group_manage.title')}</span>
          </Card.Header>
          <GroupsManager />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Group;
