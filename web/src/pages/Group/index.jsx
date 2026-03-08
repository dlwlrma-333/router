import React from 'react';
import { Card } from 'semantic-ui-react';
import GroupsManager from '../../components/GroupsManager';

const Group = () => {
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <GroupsManager />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Group;
