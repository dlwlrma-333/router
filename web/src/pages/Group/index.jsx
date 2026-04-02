import React from 'react';
import { Card } from 'semantic-ui-react';
import { useParams } from 'react-router-dom';
import GroupsManager from '../../components/GroupsManager';

const Group = () => {
  const { id: detailGroupId } = useParams();
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <GroupsManager detailGroupId={detailGroupId || ''} />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Group;
