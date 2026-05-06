import React from 'react';
import {
  Button,
  Checkbox,
  Form,
  Label,
  Message,
  Table,
} from 'semantic-ui-react';

const ChannelDetailEndpointsTab = ({
  t,
  capabilityColumnWidths,
  policyColumnWidths,
  endpointCapabilitySummaryText,
  endpointCapabilityStats,
  channelEndpoints,
  channelEndpointsLoading,
  channelEndpointsError,
  renderEndpointCapabilitySource,
  buildChannelEndpointKey,
  modelTestResultsByKey,
  endpointCapabilityReadonly,
  endpointMutatingKey,
  updateChannelEndpointCapability,
  endpointPolicySummaryText,
  endpointPolicyStats,
  channelEndpointPoliciesLoading,
  channelEndpointPolicies,
  channelEndpointPoliciesError,
  endpointPolicyReadonly,
  openEndpointPolicyEditor,
  timestamp2string,
}) => {
  return (
    <>
      <section className='router-entity-detail-section'>
        <div className='router-entity-detail-section-header'>
          <div className='router-toolbar-start router-block-gap-sm'>
            <span className='router-entity-detail-section-title'>
              {t('channel.edit.endpoint_capabilities.title')}
            </span>
            <span className='router-toolbar-meta'>
              ({endpointCapabilitySummaryText})
            </span>
          </div>
        </div>
        <Form.Field>
          <Message info className='router-section-message'>
            {t('channel.edit.endpoint_capabilities.hint')}
          </Message>
          <div className='router-detail-summary-grid'>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>
                {endpointCapabilityStats.total}
              </div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_capabilities.cards.total')}
              </div>
            </div>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>
                {endpointCapabilityStats.enabled}
              </div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_capabilities.cards.enabled')}
              </div>
            </div>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>
                {endpointCapabilityStats.explicit}
              </div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_capabilities.cards.explicit')}
              </div>
            </div>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>
                {endpointCapabilityStats.candidate}
              </div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_capabilities.cards.candidate')}
              </div>
            </div>
          </div>
          <Table
            celled
            stackable
            className='router-detail-table router-channel-endpoint-capability-table'
            compact='very'
          >
            <colgroup>
              {capabilityColumnWidths.map((width, index) => (
                <col
                  key={`channel-endpoint-capability-col-${index}`}
                  style={{ width }}
                />
              ))}
            </colgroup>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_capabilities.table.model')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_capabilities.table.endpoint')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_capabilities.table.source')}
                </Table.HeaderCell>
                <Table.HeaderCell textAlign='center'>
                  {t('channel.edit.endpoint_capabilities.table.enabled')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_capabilities.table.latest_test')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_capabilities.table.updated_at')}
                </Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {channelEndpoints.length === 0 ? (
                <Table.Row>
                  <Table.Cell className='router-empty-cell' colSpan={6}>
                    {channelEndpointsLoading
                      ? t('channel.edit.endpoint_capabilities.loading')
                      : t('channel.edit.endpoint_capabilities.empty')}
                  </Table.Cell>
                </Table.Row>
              ) : (
                channelEndpoints.map((row) => {
                  const endpointKey = buildChannelEndpointKey(
                    row.model,
                    row.endpoint,
                  );
                  const latestResult = modelTestResultsByKey.get(endpointKey) || null;
                  const latestStatusKey = latestResult
                    ? latestResult.supported === true &&
                      latestResult.status === 'supported'
                      ? 'supported'
                      : latestResult.status || 'unsupported'
                    : 'untested';
                  const isMutating = endpointMutatingKey === endpointKey;
                  return (
                    <Table.Row key={endpointKey}>
                      <Table.Cell title={row.model}>
                        <span className='router-cell-truncate'>{row.model}</span>
                      </Table.Cell>
                      <Table.Cell title={row.endpoint}>
                        <span className='router-cell-truncate'>{row.endpoint}</span>
                      </Table.Cell>
                      <Table.Cell>
                        {renderEndpointCapabilitySource(row.source)}
                      </Table.Cell>
                      <Table.Cell textAlign='center'>
                        <Checkbox
                          checked={row.enabled === true}
                          disabled={endpointCapabilityReadonly || isMutating}
                          onChange={(e, { checked }) =>
                            updateChannelEndpointCapability(row, !!checked)
                          }
                        />
                      </Table.Cell>
                      <Table.Cell
                        title={t(
                          `channel.edit.model_tester.status.${latestStatusKey}`,
                        )}
                      >
                        <span className='router-cell-truncate'>
                          {t(`channel.edit.model_tester.status.${latestStatusKey}`)}
                        </span>
                      </Table.Cell>
                      <Table.Cell className='router-nowrap'>
                        {row.updated_at > 0 ? timestamp2string(row.updated_at) : '-'}
                      </Table.Cell>
                    </Table.Row>
                  );
                })
              )}
            </Table.Body>
          </Table>
          {channelEndpointsError && (
            <div className='router-error-text router-error-text-top'>
              {channelEndpointsError}
            </div>
          )}
        </Form.Field>
      </section>
      <section className='router-entity-detail-section'>
        <div className='router-entity-detail-section-header'>
          <div className='router-toolbar-start router-block-gap-sm'>
            <span className='router-entity-detail-section-title'>
              {t('channel.edit.endpoint_policies.title')}
            </span>
            <span className='router-toolbar-meta'>({endpointPolicySummaryText})</span>
          </div>
        </div>
        <Form.Field>
          <Message info className='router-section-message'>
            {t('channel.edit.endpoint_policies.hint')}
          </Message>
          <div className='router-detail-summary-grid'>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>{endpointPolicyStats.total}</div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_policies.cards.configured')}
              </div>
            </div>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>{endpointPolicyStats.enabled}</div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_policies.cards.enabled')}
              </div>
            </div>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>{endpointPolicyStats.disabled}</div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_policies.cards.disabled')}
              </div>
            </div>
            <div className='router-inline-stat-card'>
              <div className='router-inline-stat-value'>
                {endpointPolicyStats.unconfigured}
              </div>
              <div className='router-inline-stat-hint'>
                {t('channel.edit.endpoint_policies.cards.not_configured')}
              </div>
            </div>
          </div>
          <Table
            celled
            stackable
            className='router-detail-table router-channel-endpoint-policy-table'
            compact='very'
          >
            <colgroup>
              {policyColumnWidths.map((width, index) => (
                <col
                  key={`channel-endpoint-policy-col-${index}`}
                  style={{ width }}
                />
              ))}
            </colgroup>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_policies.table.model')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_policies.table.endpoint')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_policies.table.status')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_policies.table.source')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_policies.table.reason')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_policies.table.updated_at')}
                </Table.HeaderCell>
                <Table.HeaderCell>
                  {t('channel.edit.endpoint_policies.table.actions')}
                </Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {channelEndpoints.length === 0 ? (
                <Table.Row>
                  <Table.Cell className='router-empty-cell' colSpan={7}>
                    {channelEndpointPoliciesLoading
                      ? t('channel.edit.endpoint_policies.loading')
                      : t('channel.edit.endpoint_policies.empty')}
                  </Table.Cell>
                </Table.Row>
              ) : (
                channelEndpoints.map((endpointRow) => {
                  const endpointKey = buildChannelEndpointKey(
                    endpointRow.model,
                    endpointRow.endpoint,
                  );
                  const policyRow =
                    channelEndpointPolicies.find(
                      (item) =>
                        item.model === endpointRow.model &&
                        item.endpoint === endpointRow.endpoint,
                    ) || null;
                  return (
                    <Table.Row key={`policy-${endpointKey}`}>
                      <Table.Cell title={endpointRow.model}>
                        <span className='router-cell-truncate'>
                          {endpointRow.model}
                        </span>
                      </Table.Cell>
                      <Table.Cell title={endpointRow.endpoint}>
                        <span className='router-cell-truncate'>
                          {endpointRow.endpoint}
                        </span>
                      </Table.Cell>
                      <Table.Cell>
                        {policyRow ? (
                          policyRow.enabled ? (
                            <Label basic color='green' className='router-tag'>
                              {t('channel.edit.endpoint_policies.status.enabled')}
                            </Label>
                          ) : (
                            <Label basic color='grey' className='router-tag'>
                              {t('channel.edit.endpoint_policies.status.disabled')}
                            </Label>
                          )
                        ) : (
                          <Label basic className='router-tag'>
                            {t(
                              'channel.edit.endpoint_policies.status.not_configured',
                            )}
                          </Label>
                        )}
                      </Table.Cell>
                      <Table.Cell>{policyRow?.source || '-'}</Table.Cell>
                      <Table.Cell title={(policyRow?.reason || '').toString()}>
                        <span className='router-cell-truncate'>
                          {policyRow?.reason || '-'}
                        </span>
                      </Table.Cell>
                      <Table.Cell className='router-nowrap'>
                        {policyRow?.updated_at > 0
                          ? timestamp2string(policyRow.updated_at)
                          : '-'}
                      </Table.Cell>
                      <Table.Cell collapsing>
                        <Button
                          type='button'
                          className='router-inline-button'
                          disabled={endpointPolicyReadonly}
                          onClick={() => openEndpointPolicyEditor(endpointRow)}
                        >
                          {policyRow
                            ? t('channel.edit.endpoint_policies.edit')
                            : t('channel.edit.endpoint_policies.create')}
                        </Button>
                      </Table.Cell>
                    </Table.Row>
                  );
                })
              )}
            </Table.Body>
          </Table>
          {channelEndpointPoliciesError && (
            <div className='router-error-text router-error-text-top'>
              {channelEndpointPoliciesError}
            </div>
          )}
        </Form.Field>
      </section>
    </>
  );
};

export default ChannelDetailEndpointsTab;
