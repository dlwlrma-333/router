import React from 'react';
import { Card } from 'semantic-ui-react';
import PackagesManager from '../../components/PackagesManager';

const Package = () => {
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <PackagesManager />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Package;
