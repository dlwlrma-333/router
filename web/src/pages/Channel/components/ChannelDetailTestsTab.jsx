import React, { useMemo, useState } from 'react';
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
  channelId,
  inputs,
  columnWidths,
  modelTestRows,
  modelTestTargetModels,
  detailModelMutating,
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
  resolvePreferredProviderForModel,
  normalizeChannelModelType,
}) => {
  const [providerFilter, setProviderFilter] = useState('');
  const [typeFilter, setTypeFilter] = useState('');

  if (inputs.protocol === 'proxy') {
    return null;
  }

  const rowsWithMeta = useMemo(
    () =>
      modelTestRows.map((row) => ({
        ...row,
        providerKey: resolvePreferredProviderForModel(row) || '',
        typeKey: normalizeChannelModelType(row.type),
      })),
    [modelTestRows, normalizeChannelModelType, resolvePreferredProviderForModel],
  );

  const providerOptions = useMemo(() => {
    const values = Array.from(
      new Set(rowsWithMeta.map((row) => row.providerKey).filter(Boolean)),
    ).sort((a, b) => a.localeCompare(b));
    return values.map((value) => ({
      key: value,
      value,
      text: value,
    }));
  }, [rowsWithMeta]);

  const typeOptions = useMemo(() => {
    const values = Array.from(
      new Set(rowsWithMeta.map((row) => row.typeKey).filter(Boolean)),
    ).sort((a, b) => a.localeCompare(b));
    return values.map((value) => ({
      key: value,
      value,
      text: t(`channel.model_types.${value}`),
    }));
  }, [rowsWithMeta, t]);

  const providerStorageKey = useMemo(
    () =>
      `channel-test-filter-provider:${(channelId || '').toString().trim() || 'create'}`,
    [channelId],
  );
  const typeStorageKey = useMemo(
    () =>
      `channel-test-filter-type:${(channelId || '').toString().trim() || 'create'}`,
    [channelId],
  );

  React.useEffect(() => {
    if (providerOptions.length === 0) {
      if (providerFilter !== '') {
        setProviderFilter('');
      }
      return;
    }
    const validValues = new Set(providerOptions.map((item) => item.value));
    if (providerFilter !== '' && validValues.has(providerFilter)) {
      return;
    }
    const storedValue = window.localStorage.getItem(providerStorageKey) || '';
    const nextValue = validValues.has(storedValue)
      ? storedValue
      : providerOptions[0]?.value || '';
    if (nextValue !== providerFilter) {
      setProviderFilter(nextValue);
    }
  }, [providerFilter, providerOptions, providerStorageKey]);

  React.useEffect(() => {
    if (typeOptions.length === 0) {
      if (typeFilter !== '') {
        setTypeFilter('');
      }
      return;
    }
    const validValues = new Set(typeOptions.map((item) => item.value));
    if (typeFilter !== '' && validValues.has(typeFilter)) {
      return;
    }
    const storedValue = window.localStorage.getItem(typeStorageKey) || '';
    const nextValue = validValues.has(storedValue)
      ? storedValue
      : typeOptions[0]?.value || '';
    if (nextValue !== typeFilter) {
      setTypeFilter(nextValue);
    }
  }, [typeFilter, typeOptions, typeStorageKey]);

  const filteredRows = useMemo(
    () =>
      rowsWithMeta.filter((row) => {
        if (providerFilter !== '' && row.providerKey !== providerFilter) {
          return false;
        }
        if (typeFilter !== '' && row.typeKey !== typeFilter) {
          return false;
        }
        return true;
      }),
    [providerFilter, rowsWithMeta, typeFilter],
  );

  const filteredModelIDs = useMemo(
    () => filteredRows.map((row) => row.model),
    [filteredRows],
  );
  const filteredTargetSet = useMemo(
    () => new Set(modelTestTargetModels),
    [modelTestTargetModels],
  );
  const filteredAllSelected =
    filteredRows.length > 0 &&
    filteredRows.every((row) => filteredTargetSet.has(row.model));
  const filteredPartiallySelected =
    !filteredAllSelected &&
    filteredRows.some((row) => filteredTargetSet.has(row.model));
  const filteredSelectedCount = filteredRows.filter((row) =>
    filteredTargetSet.has(row.model),
  ).length;

  const batchEndpointOptions = useMemo(() => {
    const map = new Map();
    filteredRows.forEach((row) => {
      getEndpointOptionsForModel(row).forEach((option) => {
        if (!map.has(option.value)) {
          map.set(option.value, option);
        }
      });
    });
    return Array.from(map.values());
  }, [filteredRows, getEndpointOptionsForModel]);

  const batchEndpointValue = useMemo(() => {
    const endpointSet = new Set(
      filteredRows.map((row) => getEffectiveModelEndpoint(row)).filter(Boolean),
    );
    return endpointSet.size === 1 ? Array.from(endpointSet)[0] || '' : '';
  }, [filteredRows, getEffectiveModelEndpoint]);

  const disabledBase = detailTestingReadonly || detailModelMutating;

  const toggleFilteredTargets = (checked) => {
    const targetSet = new Set(filteredTargetSet);
    filteredModelIDs.forEach((model) => {
      if (checked) {
        targetSet.add(model);
      } else {
        targetSet.delete(model);
      }
    });
    const nextSelected = Array.from(targetSet);
    modelTestRows.forEach((row) => {
      const shouldSelect = nextSelected.includes(row.model);
      const isSelected = filteredTargetSet.has(row.model);
      if (shouldSelect !== isSelected) {
        toggleModelTestTarget(row.model, shouldSelect);
      }
    });
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
        <div className='router-toolbar router-block-gap-sm'>
          <div className='router-toolbar-start router-block-gap-sm'>
            <Dropdown
              selection
              className='router-section-dropdown router-detail-filter-dropdown router-dropdown-min-170'
              options={providerOptions}
              value={providerFilter || undefined}
              disabled={disabledBase || providerOptions.length === 0}
              placeholder={t('channel.edit.model_tester.filters.provider')}
              onChange={(e, { value }) =>
                {
                  const nextValue = (value || '').toString();
                  setProviderFilter(nextValue);
                  if (nextValue !== '') {
                    window.localStorage.setItem(providerStorageKey, nextValue);
                  }
                }
              }
            />
            <Dropdown
              selection
              className='router-section-dropdown router-detail-filter-dropdown router-dropdown-min-170'
              options={typeOptions}
              value={typeFilter || undefined}
              disabled={disabledBase || typeOptions.length === 0}
              placeholder={t('channel.edit.model_tester.filters.type')}
              onChange={(e, { value }) =>
                {
                  const nextValue = (value || '').toString();
                  setTypeFilter(nextValue);
                  if (nextValue !== '') {
                    window.localStorage.setItem(typeStorageKey, nextValue);
                  }
                }
              }
            />
            <Dropdown
              selection
              clearable
              className='router-section-dropdown router-detail-filter-dropdown router-dropdown-min-170'
              options={batchEndpointOptions}
              value={batchEndpointValue || undefined}
              disabled={disabledBase || batchEndpointOptions.length === 0}
              placeholder={t('channel.edit.model_tester.table.batch_set')}
              onChange={(e, { value }) => {
                if ((value || '').toString().trim() === '') {
                  return;
                }
                updateAllModelTestEndpoints(value, filteredModelIDs);
              }}
            />
          </div>
          <div className='router-toolbar-end router-block-gap-sm'>
            <Button
              type='button'
              className='router-section-button'
              color='blue'
              loading={modelTesting && modelTestingScope === 'batch'}
              disabled={
                disabledBase ||
                modelTesting ||
                filteredSelectedCount === 0 ||
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
                selected: filteredSelectedCount,
                total: filteredRows.length,
              })}
            </Label>
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
        <Table
          celled
          stackable
          className='router-detail-table router-model-test-table'
        >
          <colgroup>
            {columnWidths.map((width, index) => (
              <col
                key={`channel-model-test-col-${index}`}
                style={{ width }}
              />
            ))}
          </colgroup>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell collapsing textAlign='center'>
                <Checkbox
                  checked={filteredAllSelected}
                  indeterminate={filteredPartiallySelected}
                  disabled={disabledBase || filteredRows.length === 0}
                  onChange={(e, { checked }) =>
                    toggleFilteredTargets(!!checked)
                  }
                />
              </Table.HeaderCell>
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
            {filteredRows.length === 0 ? (
              <Table.Row>
                <Table.Cell className='router-empty-cell' colSpan='9'>
                  {t(
                    modelTestRows.length === 0
                      ? 'channel.edit.model_tester.empty'
                      : 'channel.edit.model_selector.empty_filtered',
                  )}
                </Table.Cell>
              </Table.Row>
            ) : (
              filteredRows.map((row) => {
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
              })
            )}
          </Table.Body>
        </Table>
      </Form.Field>
    </section>
  );
};

export default ChannelDetailTestsTab;
