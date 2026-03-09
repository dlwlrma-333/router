import React from 'react';
import {Card} from 'semantic-ui-react';
import ChannelsTable from '../../components/ChannelsTable';

const Channel = () => {
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <ChannelsTable />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Channel;
