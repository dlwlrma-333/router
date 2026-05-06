import React from 'react';
import {
  Button,
  Checkbox,
  Dropdown,
  Form,
  Label,
  Message,
  Table,
} from 'semantic-ui-react';

const ChannelDetailTestsTab = ({
  t,
  inputs,
  columnWidths,
  modelTestRows,
  modelTestGroups,
  modelTestTargetModels,
  detailModelMutating,
  toggleModelTestGroupTargets,
  toggleModelTestTarget,
  getEffectiveModelEndpoint,
  modelTestResultsByKey,
  buildModelTestResultKey,
  latestModelTestResultByModel,
  activeChannelTasksByModel,
  getEndpointOptionsForModel,
  updateModelTestEndpoint,
  updateModelTestStream,
  modelTesting,
  modelTestingScope,
  modelTestingTargetSet,
  handleRunModelTests,
  handleDownloadModelTestArtifact,
  detailTestingReadonly,
  modelTestError,
  openChannelTaskView,
  modelTestedAt,
  selectedModelTestHasActiveTasks,
  timestamp2string,
  updateAllModelTestEndpoints,
}) => {
  if (inputs.protocol === 'proxy') {
    return null;
  }

  const renderModelTestGroupTable = (group, readonly = false) => {
    const rows = Array.isArray(group?.rows) ? group.rows : [];
    const groupModelIDs = rows.map((row) => row.model);
    const groupAllSelected =
      rows.length > 0 &&
      rows.every((row) => modelTestTargetModels.includes(row.model));
    const groupPartiallySelected =
      !groupAllSelected &&
      rows.some((row) => modelTestTargetModels.includes(row.model));
    const disabledBase = readonly || detailModelMutating;
    return (
      <div className='router-block-gap-md' key={group.key}>
        <div className='router-toolbar router-block-gap-xs'>
          <div className='router-toolbar-start'>
            <Label basic className='router-tag'>
              {group.provider || '-'}
            </Label>
            <Label basic className='router-tag'>
              {t('channel.edit.model_tester.group_type', {
                type: t(`channel.model_types.${group.type || 'text'}`),
              })}
            </Label>
            <span className='router-toolbar-meta'>
              {t('channel.edit.model_tester.group_count', {
                count: rows.length,
              })}
            </span>
          </div>
        </div>
        <Table
          celled
          stackable
          className='router-detail-table router-model-test-table'
        >
          <colgroup>
            {columnWidths.map((width, index) => (
              <col
                key={`channel-model-test-group-col-${index}`}
                style={{ width }}
              />
            ))}
          </colgroup>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell collapsing textAlign='center'>
                <Checkbox
                  checked={groupAllSelected}
                  indeterminate={groupPartiallySelected}
                  disabled={disabledBase}
                  onChange={(e, { checked }) =>
                    toggleModelTestGroupTargets(rows, !!checked)
                  }
                />
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_tester.table.model')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                <div className='router-table-header-with-control'>
                  <span>{t('channel.edit.model_tester.table.endpoint')}</span>
                  <Dropdown
                    selection
                    compact
                    className='router-inline-dropdown'
                    placeholder={t('channel.edit.model_tester.table.batch_set')}
                    options={group.endpointOptions}
                    disabled={disabledBase || group.endpointOptions.length === 0}
                    value={group.endpointValue || undefined}
                    onClick={(e) => {
                      e.stopPropagation();
                    }}
                    onChange={(e, { value }) => {
                      updateAllModelTestEndpoints(value, groupModelIDs);
                    }}
                  />
                </div>
              </Table.HeaderCell>
              <Table.HeaderCell collapsing>
                {t('channel.edit.model_tester.table.is_stream')}
              </Table.HeaderCell>
              <Table.HeaderCell collapsing>
                {t('channel.edit.model_tester.table.status')}
              </Table.HeaderCell>
              <Table.HeaderCell collapsing>
                {t('channel.edit.model_tester.table.latency')}
              </Table.HeaderCell>
              <Table.HeaderCell collapsing>
                {t('channel.edit.model_tester.table.tested_at')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_tester.table.message')}
              </Table.HeaderCell>
              <Table.HeaderCell collapsing>
                {t('channel.edit.model_tester.table.actions')}
              </Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {rows.map((row) => {
              const normalizedEndpoint = getEffectiveModelEndpoint(row);
              const item = modelTestResultsByKey.get(
                buildModelTestResultKey(row.model, normalizedEndpoint),
              );
              const latestItemForModel =
                latestModelTestResultByModel.get(row.model) || null;
              const displayItem = item || latestItemForModel;
              const activeTask = activeChannelTasksByModel.get(row.model) || null;
              const useLatestResult = !activeTask && !item && !!latestItemForModel;
              const canDownloadArtifact =
                !!displayItem?.artifact_path || !!displayItem?.artifact_name;
              const effectiveStatus =
                activeTask?.status || displayItem?.status || 'untested';
              const labelColor =
                effectiveStatus === 'running'
                  ? 'blue'
                  : effectiveStatus === 'pending'
                    ? 'orange'
                    : effectiveStatus === 'untested'
                      ? undefined
                      : effectiveStatus === 'supported'
                        ? 'green'
                        : effectiveStatus === 'skipped'
                          ? 'grey'
                          : 'red';
              return (
                <Table.Row key={row.model}>
                  <Table.Cell textAlign='center'>
                    <Checkbox
                      checked={modelTestTargetModels.includes(row.model)}
                      disabled={disabledBase}
                      onChange={(e, { checked }) =>
                        toggleModelTestTarget(row.model, !!checked)
                      }
                    />
                  </Table.Cell>
                  <Table.Cell title={row.model || '-'}>
                    <span className='router-cell-truncate'>{row.model || '-'}</span>
                  </Table.Cell>
                  <Table.Cell className='router-table-dropdown-cell'>
                    {row.type === 'text' || row.type === 'image' ? (
                      <Dropdown
                        selection
                        className='router-mini-dropdown router-table-dropdown-fluid'
                        options={getEndpointOptionsForModel(row)}
                        disabled={disabledBase}
                        value={normalizedEndpoint}
                        onChange={(e, { value }) =>
                          updateModelTestEndpoint(row.model, value)
                        }
                      />
                    ) : (
                      normalizedEndpoint || row.endpoint || '-'
                    )}
                  </Table.Cell>
                  <Table.Cell textAlign='center'>
                    <Checkbox
                      checked={!!row.is_stream}
                      disabled={disabledBase}
                      onChange={(e, { checked }) =>
                        updateModelTestStream(row.model, !!checked)
                      }
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Label basic color={labelColor} className='router-tag'>
                      {t(`channel.edit.model_tester.status.${effectiveStatus}`)}
                    </Label>
                  </Table.Cell>
                  <Table.Cell className='router-nowrap'>
                    {displayItem?.latency_ms > 0
                      ? `${displayItem.latency_ms} ms`
                      : '-'}
                  </Table.Cell>
                  <Table.Cell className='router-nowrap'>
                    {displayItem?.tested_at > 0
                      ? timestamp2string(displayItem.tested_at)
                      : '-'}
                  </Table.Cell>
                  <Table.Cell
                    title={
                      useLatestResult
                        ? t(
                            'channel.edit.model_tester.latest_result_from_endpoint',
                            {
                              endpoint: displayItem?.endpoint || '-',
                            },
                          )
                        : displayItem?.message ||
                          (effectiveStatus === 'untested'
                            ? t('channel.edit.model_tester.untested')
                            : '-')
                    }
                  >
                    <span className='router-cell-truncate'>
                      {useLatestResult
                        ? t(
                            'channel.edit.model_tester.latest_result_from_endpoint',
                            {
                              endpoint: displayItem?.endpoint || '-',
                            },
                          )
                        : displayItem?.message ||
                          (effectiveStatus === 'untested'
                            ? t('channel.edit.model_tester.untested')
                            : '-')}
                    </span>
                  </Table.Cell>
                  <Table.Cell collapsing>
                    <div className='router-inline-actions'>
                      <Button
                        type='button'
                        className='router-inline-button'
                        basic
                        loading={
                          (modelTesting &&
                            modelTestingScope === 'single' &&
                            modelTestingTargetSet.has(row.model)) ||
                          !!activeTask
                        }
                        disabled={
                          disabledBase ||
                          modelTesting ||
                          activeChannelTasksByModel.has(row.model)
                        }
                        onClick={() =>
                          handleRunModelTests({
                            targetModels: [row.model],
                            scope: 'single',
                          })
                        }
                      >
                        {t('channel.edit.model_tester.single')}
                      </Button>
                      <Button
                        type='button'
                        className='router-inline-button'
                        basic
                        disabled={!canDownloadArtifact}
                        onClick={() => handleDownloadModelTestArtifact(displayItem)}
                      >
                        {t('common.download')}
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
          </Table.Body>
        </Table>
      </div>
    );
  };

  const renderModelTestGroups = (readonly = false, groups = modelTestGroups) => {
    if (modelTestRows.length === 0 || groups.length === 0) {
      return (
        <Table celled stackable className='router-detail-table'>
          <Table.Body>
            <Table.Row>
              <Table.Cell className='router-empty-cell' colSpan='10'>
                {t(
                  modelTestRows.length === 0
                    ? 'channel.edit.model_tester.empty'
                    : 'channel.edit.model_selector.empty_filtered',
                )}
              </Table.Cell>
            </Table.Row>
          </Table.Body>
        </Table>
      );
    }
    return groups.map((group) => renderModelTestGroupTable(group, readonly));
  };

  return (
    <section className='router-entity-detail-section'>
      <div className='router-entity-detail-section-header'>
        <div className='router-toolbar-start'>
          <span className='router-entity-detail-section-title'>
            {t('channel.edit.model_tester.title')}
          </span>
        </div>
      </div>
      <Form.Field>
        <Message info className='router-section-message'>
          {t('channel.edit.model_tester.hint')}
        </Message>
        <Message className='router-section-message'>
          {t('channel.edit.model_tester.selection_notice')}
        </Message>
        <div className='router-toolbar-end router-block-gap-sm'>
          <>
            <Button
              type='button'
              className='router-section-button'
              color='blue'
              loading={modelTesting && modelTestingScope === 'batch'}
              disabled={
                detailTestingReadonly ||
                detailModelMutating ||
                modelTesting ||
                modelTestTargetModels.length === 0 ||
                selectedModelTestHasActiveTasks
              }
              onClick={() =>
                handleRunModelTests({
                  targetModels: modelTestTargetModels,
                  scope: 'batch',
                })
              }
            >
              {t('channel.edit.model_tester.button')}
            </Button>
            <Label basic className='router-tag'>
              {t('channel.edit.model_tester.selection', {
                selected: modelTestTargetModels.length,
                total: modelTestRows.length,
              })}
            </Label>
          </>
          <Button
            type='button'
            className='router-page-button'
            basic
            onClick={() =>
              openChannelTaskView({
                type: 'channel_model_test',
              })
            }
          >
            {t('channel.edit.model_tester.history_tasks')}
          </Button>
          {modelTestedAt > 0 && (
            <span className='router-toolbar-meta'>
              {t('channel.edit.model_tester.last_tested', {
                time: new Date(modelTestedAt).toLocaleString(),
              })}
            </span>
          )}
        </div>
        {modelTestError && (
          <div className='router-error-text router-block-gap-sm'>
            {modelTestError}
          </div>
        )}
        {renderModelTestGroups(true)}
      </Form.Field>
    </section>
  );
};

export default ChannelDetailTestsTab;
