import React from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Button, Card, Label, Statistic, Table } from 'semantic-ui-react';
import { timestamp2string } from '../../helpers';
import RedeemCodePage from './RedeemCodePage';
import {
  renderTopupIntegerAmountWithExactPopup,
  useTopUpWorkspace,
} from './shared.jsx';

const formatLotSource = (source, t) => {
  switch ((source || '').trim()) {
    case 'topup_order':
      return t('topup.balance_lots.source.topup_order');
    case 'redemption':
      return t('topup.balance_lots.source.redemption');
    case 'legacy_migration':
      return t('topup.balance_lots.source.legacy_migration');
    default:
      return source || '-';
  }
};

const renderLotStatus = (status, t) => {
  switch ((status || '').trim()) {
    case 'active':
      return (
        <Label basic color='green' className='router-tag'>
          {t('topup.balance_lots.status.active')}
        </Label>
      );
    case 'exhausted':
      return (
        <Label basic color='grey' className='router-tag'>
          {t('topup.balance_lots.status.exhausted')}
        </Label>
      );
    case 'expired':
      return (
        <Label basic color='orange' className='router-tag'>
          {t('topup.balance_lots.status.expired')}
        </Label>
      );
    default:
      return <Label basic className='router-tag'>{status || '-'}</Label>;
  }
};

const BalanceStatusPage = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const {
    userBalanceYYC,
    topupBalanceYYC,
    redeemBalanceYYC,
    balanceLots,
    loadingBalanceLots,
    loadBalanceLots,
    displayCurrency,
    displayCurrencyIndex,
  } = useTopUpWorkspace();

  return (
    <div style={{ display: 'grid', gap: '1rem' }}>
      <Card fluid className='router-soft-card router-soft-card-fill'>
        <Card.Content className='router-card-fill'>
          <Card.Description className='router-card-fill'>
            <div className='router-card-body-spread'>
              <div
                style={{
                  display: 'grid',
                  gap: '1rem',
                  gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
                  alignItems: 'center',
                }}
              >
                <div className='router-center-panel' style={{ paddingTop: 0 }}>
                  <Statistic className='router-accent-statistic' size='small'>
                    <Statistic.Value className='router-topup-statistic-value'>
                      {renderTopupIntegerAmountWithExactPopup({
                        yycAmount: userBalanceYYC,
                        displayCurrency,
                        displayCurrencyIndex,
                      })}
                    </Statistic.Value>
                    <Statistic.Label>
                      {t('topup.external_topup.total_balance')}
                    </Statistic.Label>
                  </Statistic>
                </div>
                <div className='router-center-panel' style={{ paddingTop: 0 }}>
                  <Statistic size='small'>
                    <Statistic.Value className='router-topup-statistic-value'>
                      {renderTopupIntegerAmountWithExactPopup({
                        yycAmount: topupBalanceYYC,
                        displayCurrency,
                        displayCurrencyIndex,
                      })}
                    </Statistic.Value>
                    <Statistic.Label>
                      {t('topup.external_topup.topup_balance')}
                    </Statistic.Label>
                  </Statistic>
                </div>
                <div className='router-center-panel' style={{ paddingTop: 0 }}>
                  <Statistic size='small'>
                    <Statistic.Value className='router-topup-statistic-value'>
                      {renderTopupIntegerAmountWithExactPopup({
                        yycAmount: redeemBalanceYYC,
                        displayCurrency,
                        displayCurrencyIndex,
                      })}
                    </Statistic.Value>
                    <Statistic.Label>
                      {t('topup.external_topup.redeem_balance')}
                    </Statistic.Label>
                  </Statistic>
                </div>
              </div>
              <div className='router-action-footer'>
                <Button
                  primary
                  fluid
                  className='router-section-button'
                  onClick={() => navigate('/workspace/service/pricing')}
                >
                  {t('topup.record_nav.topup')}
                </Button>
              </div>
            </div>
          </Card.Description>
        </Card.Content>
      </Card>
      <RedeemCodePage />
      <Card fluid className='router-soft-card'>
        <Card.Content>
          <div className='router-toolbar router-block-gap-sm' style={{ marginBottom: '0.75rem' }}>
            <div className='router-toolbar-start'>
              <div className='router-section-title' style={{ margin: 0 }}>
                {t('topup.balance_lots.title')}
              </div>
            </div>
            <div className='router-toolbar-end'>
              <Button
                className='router-section-button'
                loading={loadingBalanceLots}
                onClick={() => loadBalanceLots()}
              >
                {t('common.refresh')}
              </Button>
            </div>
          </div>
          {balanceLots.length === 0 ? (
            <div className='router-empty'>{t('topup.balance_lots.empty')}</div>
          ) : (
            <div className='router-table-scroll-x'>
              <Table celled className='router-table router-list-table'>
                <Table.Header>
                  <Table.Row>
                    <Table.HeaderCell>{t('topup.balance_lots.columns.source')}</Table.HeaderCell>
                    <Table.HeaderCell>{t('topup.balance_lots.columns.remaining')}</Table.HeaderCell>
                    <Table.HeaderCell>{t('topup.balance_lots.columns.total')}</Table.HeaderCell>
                    <Table.HeaderCell>{t('topup.balance_lots.columns.status')}</Table.HeaderCell>
                    <Table.HeaderCell>{t('topup.balance_lots.columns.granted_at')}</Table.HeaderCell>
                    <Table.HeaderCell>{t('topup.balance_lots.columns.expires_at')}</Table.HeaderCell>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {balanceLots.map((row) => (
                    <Table.Row key={row.id || `${row.source_type}-${row.source_id}`}>
                      <Table.Cell>{formatLotSource(row.source_type, t)}</Table.Cell>
                      <Table.Cell>
                        {renderTopupIntegerAmountWithExactPopup({
                          yycAmount: row.remaining_yyc,
                          displayCurrency,
                          displayCurrencyIndex,
                        })}
                      </Table.Cell>
                      <Table.Cell>
                        {renderTopupIntegerAmountWithExactPopup({
                          yycAmount: row.total_yyc,
                          displayCurrency,
                          displayCurrencyIndex,
                        })}
                      </Table.Cell>
                      <Table.Cell>{renderLotStatus(row.status, t)}</Table.Cell>
                      <Table.Cell>
                        {row.granted_at ? timestamp2string(row.granted_at) : '-'}
                      </Table.Cell>
                      <Table.Cell>
                        {Number(row.expires_at || 0) > 0
                          ? timestamp2string(row.expires_at)
                          : t('common.never')}
                      </Table.Cell>
                    </Table.Row>
                  ))}
                </Table.Body>
              </Table>
            </div>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default BalanceStatusPage;
