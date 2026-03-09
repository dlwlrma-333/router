import React from 'react';
import { Card } from 'semantic-ui-react';
import RedemptionsTable from '../../components/RedemptionsTable';

const Redemption = () => {
  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <RedemptionsTable />
        </Card.Content>
      </Card>
    </div>
  );
};

export default Redemption;
