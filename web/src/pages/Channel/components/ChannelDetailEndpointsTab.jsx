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
  columnWidths,
  endpointSummaryText,
  channelEndpoints,
  channelEndpointsLoading,
  channelEndpointsError,
  buildChannelEndpointKey,
  modelTestResultsByKey,
  endpointCapabilityReadonly,
  endpointMutatingKey,
  updateChannelEndpointCapability,
  channelEndpointPoliciesLoading,
  channelEndpointPolicies,
  channelEndpointPoliciesError,
  endpointPolicyReadonly,
  openEndpointPolicyEditor,
  timestamp2string,
}) => {
  const policyByKey = new Map(
    channelEndpointPolicies.map((row) => [
      buildChannelEndpointKey(row.model, row.endpoint),
      row,
    ]),
  );
  return (
    <section className='router-entity-detail-section'>
      <div className='router-entity-detail-section-header'>
        <div className='router-toolbar-start router-block-gap-sm'>
          <span className='router-entity-detail-section-title'>
            {t('channel.edit.endpoint_capabilities.title')}
          </span>
          <span className='router-toolbar-meta'>({endpointSummaryText})</span>
        </div>
      </div>
      <Form.Field>
        <Message info className='router-section-message'>
          {t('channel.edit.endpoint_capabilities.hint')}
        </Message>
        <Table
          celled
          stackable
          className='router-detail-table router-channel-endpoint-capability-table'
          compact='very'
        >
          <colgroup>
            {columnWidths.map((width, index) => (
              <col
                key={`channel-endpoint-col-${index}`}
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
              <Table.HeaderCell textAlign='center'>
                {t('channel.edit.endpoint_capabilities.table.enabled')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.endpoint_capabilities.table.test_status')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.endpoint_policies.table.policy')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.endpoint_policies.table.policy_note')}
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
                const policyRow = policyByKey.get(endpointKey) || null;
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
                    <Table.Cell>
                      {channelEndpointPoliciesLoading &&
                      channelEndpointPolicies.length === 0 ? (
                        <Label basic className='router-tag'>
                          {t('channel.edit.endpoint_policies.loading')}
                        </Label>
                      ) : policyRow ? (
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
                          {t('channel.edit.endpoint_policies.status.not_configured')}
                        </Label>
                      )}
                    </Table.Cell>
                    <Table.Cell title={(policyRow?.reason || '').toString()}>
                      <span className='router-cell-truncate'>
                        {policyRow?.reason || '-'}
                      </span>
                    </Table.Cell>
                    <Table.Cell collapsing>
                      <Button
                        type='button'
                        className='router-inline-button'
                        disabled={endpointPolicyReadonly}
                        onClick={() => openEndpointPolicyEditor(row)}
                        title={
                          policyRow?.updated_at > 0
                            ? timestamp2string(policyRow.updated_at)
                            : row.updated_at > 0
                              ? timestamp2string(row.updated_at)
                              : undefined
                        }
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
        {channelEndpointsError && (
          <div className='router-error-text router-error-text-top'>
            {channelEndpointsError}
          </div>
        )}
        {channelEndpointPoliciesError && (
          <div className='router-error-text router-error-text-top'>
            {channelEndpointPoliciesError}
          </div>
        )}
      </Form.Field>
    </section>
  );
};

export default ChannelDetailEndpointsTab;
