import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Checkbox, Dropdown, Label, Message, Table } from 'semantic-ui-react';
import { timestamp2string } from '../../helpers';

const MODEL_TEST_COLUMN_WIDTHS = ['16%', '20%', '6%', '8%', '9%', '14%', '15%', '12%'];
const MODEL_TEST_BATCH_COLUMN_WIDTHS = ['4%', '15%', '22%', '6%', '8%', '9%', '14%', '12%', '12%'];

const CreateChannelModelTestSection = ({
  modelTesting,
  modelTestingScope,
  detailModelMutating,
  modelTestTargetModels,
  selectedModelTestHasActiveTasks,
  handleRunModelTests,
  createModelTestProviderOptions,
  createModelTestProviderFilter,
  setCreateModelTestProviderFilter,
  createModelTestTypeOptions,
  createModelTestTypeFilter,
  setCreateModelTestTypeFilter,
  createModelTestBulkEndpointOptions,
  createModelTestBulkEndpointValue,
  updateAllModelTestEndpoints,
  openChannelTaskView,
  modelTestedAt,
  modelTestError,
  modelTestRows,
  createFilteredModelTestRows,
  modelTestingTargetSet,
  activeChannelTasksByModel,
  modelTestResultsByKey,
  latestModelTestResultByModel,
  buildModelTestResultKey,
  getEffectiveModelEndpoint,
  getEndpointOptionsForModel,
  updateModelTestEndpoint,
  updateModelTestStream,
  toggleModelTestGroupTargets,
  toggleModelTestTarget,
  handleDownloadModelTestArtifact,
}) => {
  const { t } = useTranslation();
  const [batchMode, setBatchMode] = useState(false);

  const renderEmptyState = () => (
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

  const renderRowsTable = () => {
    const rows = Array.isArray(createFilteredModelTestRows)
      ? createFilteredModelTestRows
      : [];
    const disabledBase = detailModelMutating;
    const allSelected =
      rows.length > 0 &&
      rows.every((row) => modelTestTargetModels.includes(row.model));
    const partiallySelected =
      !allSelected &&
      rows.some((row) => modelTestTargetModels.includes(row.model));

    return (
      <Table celled stackable className='router-detail-table router-model-test-table'>
        <colgroup>
          {(batchMode
            ? MODEL_TEST_BATCH_COLUMN_WIDTHS
            : MODEL_TEST_COLUMN_WIDTHS
          ).map((width, index) => (
            <col key={`model-test-col-${index}`} style={{ width }} />
          ))}
        </colgroup>
        <Table.Header>
          <Table.Row>
            {batchMode && (
              <Table.HeaderCell collapsing textAlign='center'>
                <Checkbox
                  checked={allSelected}
                  indeterminate={partiallySelected}
                  disabled={disabledBase}
                  onChange={(e, { checked }) =>
                    toggleModelTestGroupTargets(rows, !!checked)
                  }
                />
              </Table.HeaderCell>
            )}
            <Table.HeaderCell>
              {t('channel.edit.model_tester.table.model')}
            </Table.HeaderCell>
            <Table.HeaderCell>
              {t('channel.edit.model_tester.table.endpoint')}
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
                {batchMode && (
                  <Table.Cell textAlign='center'>
                    <Checkbox
                      checked={modelTestTargetModels.includes(row.model)}
                      disabled={disabledBase}
                      onChange={(e, { checked }) =>
                        toggleModelTestTarget(row.model, !!checked)
                      }
                    />
                  </Table.Cell>
                )}
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
                  {displayItem?.latency_ms > 0 ? `${displayItem.latency_ms} ms` : '-'}
                </Table.Cell>
                <Table.Cell className='router-nowrap'>
                  {displayItem?.tested_at > 0
                    ? timestamp2string(displayItem.tested_at)
                    : '-'}
                </Table.Cell>
                <Table.Cell
                  title={
                    useLatestResult
                      ? t('channel.edit.model_tester.latest_result_from_endpoint', {
                          endpoint: displayItem?.endpoint || '-',
                        })
                      : displayItem?.message ||
                        (effectiveStatus === 'untested'
                          ? t('channel.edit.model_tester.untested')
                          : '-')
                  }
                >
                  <span className='router-cell-truncate'>
                    {useLatestResult
                      ? t('channel.edit.model_tester.latest_result_from_endpoint', {
                          endpoint: displayItem?.endpoint || '-',
                        })
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
    );
  };

  return (
    <>
      <div className='router-toolbar router-block-gap-xs'>
        <div className='router-toolbar-start'>
          <span className='router-entity-detail-section-title'>
            {t('channel.edit.model_tester.title')}
          </span>
        </div>
      </div>
      <Message info className='router-section-message'>
        {t('channel.edit.model_tester.hint')}
      </Message>
      <Message className='router-section-message'>
        {t('channel.edit.model_tester.selection_notice')}
      </Message>
      <div className='router-toolbar router-block-gap-sm'>
        <div className='router-toolbar-start router-block-gap-sm'>
          <Dropdown
            selection
            compact
            className='router-inline-dropdown router-inline-dropdown-fixed'
            placeholder={t('channel.edit.model_tester.filters.provider')}
            options={createModelTestProviderOptions}
            value={createModelTestProviderFilter || undefined}
            onChange={(e, { value }) =>
              setCreateModelTestProviderFilter((value || '').toString())
            }
          />
          <Dropdown
            selection
            compact
            className='router-inline-dropdown router-inline-dropdown-fixed'
            placeholder={t('channel.edit.model_tester.filters.type')}
            options={createModelTestTypeOptions}
            value={createModelTestTypeFilter || undefined}
            onChange={(e, { value }) =>
              setCreateModelTestTypeFilter((value || '').toString())
            }
          />
          <Dropdown
            selection
            compact
            className='router-inline-dropdown router-inline-dropdown-fixed'
            placeholder={t('channel.edit.model_tester.table.endpoint')}
            options={createModelTestBulkEndpointOptions}
            value={createModelTestBulkEndpointValue || undefined}
            disabled={
              createModelTestProviderFilter === '' ||
              createModelTestTypeFilter === '' ||
              createFilteredModelTestRows.length === 0 ||
              createModelTestBulkEndpointOptions.length === 0
            }
            onChange={(e, { value }) =>
              updateAllModelTestEndpoints(
                value,
                createFilteredModelTestRows.map((row) => row.model),
              )
            }
          />
        </div>
        <div className='router-toolbar-end router-block-gap-sm'>
          {batchMode && (
            <span className='router-toolbar-meta router-toolbar-meta-soft'>
              {t('channel.edit.model_tester.selection', {
                selected: modelTestTargetModels.length,
                total: modelTestRows.length,
              })}
            </span>
          )}
          <Button
            type='button'
            className='router-section-button'
            color='blue'
            loading={modelTesting && modelTestingScope === 'batch'}
            disabled={
              detailModelMutating ||
              modelTesting ||
              (batchMode && modelTestTargetModels.length === 0) ||
              selectedModelTestHasActiveTasks
            }
            onClick={() => {
              if (!batchMode) {
                setBatchMode(true);
                return;
              }
              handleRunModelTests({
                targetModels: modelTestTargetModels,
                scope: 'batch',
              });
            }}
          >
            {batchMode
              ? t('channel.edit.model_tester.button_run_batch')
              : t('channel.edit.model_tester.button_enter_batch')}
          </Button>
          {batchMode && (
            <Button
              type='button'
              className='router-page-button'
              basic
              onClick={() => {
                setBatchMode(false);
                toggleModelTestGroupTargets(createFilteredModelTestRows, false);
              }}
            >
              {t('common.cancel')}
            </Button>
          )}
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
      </div>
      {modelTestError && (
        <div className='router-error-text router-block-gap-sm'>
          {modelTestError}
        </div>
      )}
      {createFilteredModelTestRows.length === 0
        ? renderEmptyState()
        : renderRowsTable()}
    </>
  );
};

export default CreateChannelModelTestSection;
