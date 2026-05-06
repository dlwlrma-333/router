import React from 'react';
import {
  Button,
  Checkbox,
  Dropdown,
  Form,
  Label,
  Message,
  Pagination,
  Table,
} from 'semantic-ui-react';

const ChannelDetailModelsTab = ({
  t,
  columnWidths,
  modelSectionMetaText,
  detailModelFilter,
  setDetailModelFilter,
  detailModelsEditing,
  modelSearchKeyword,
  setModelSearchKeyword,
  fetchModelsLoading,
  activeRefreshModelsTask,
  detailModelMutating,
  handleFetchModels,
  searchedModelConfigs,
  visibleModelConfigs,
  renderedModelConfigs,
  getComplexPricingDetailsForModel,
  openComplexPricingModal,
  detailModelsEditLocked,
  providerCatalogLoading,
  renderModelToggleCells,
  canSelectChannelModel,
  detailCurrentPageAllSelected,
  detailCurrentPagePartiallySelected,
  detailCurrentPageSelectableCount,
  toggleDetailCurrentPageSelections,
  normalizeChannelModelType,
  startDetailModelEdit,
  detailModelTotalPages,
  detailModelPage,
  setDetailModelPage,
  modelsSyncError,
}) => {
  return (
    <section className='router-entity-detail-section'>
      <div className='router-entity-detail-section-header'>
        <div className='router-toolbar-start router-block-gap-sm'>
          <span className='router-entity-detail-section-title'>
            {t('channel.edit.detail_models_title')}
          </span>
          <span className='router-toolbar-meta'>({modelSectionMetaText})</span>
        </div>
        <div className='router-toolbar-end router-block-gap-sm'>
          <Dropdown
            selection
            className='router-section-dropdown router-dropdown-min-170 router-detail-filter-dropdown'
            compact
            disabled={detailModelsEditing}
            options={[
              {
                key: 'all',
                value: 'all',
                text: t('channel.edit.model_selector.filters.all'),
              },
              {
                key: 'enabled',
                value: 'enabled',
                text: t('channel.edit.model_selector.filters.enabled'),
              },
              {
                key: 'disabled',
                value: 'disabled',
                text: t('channel.edit.model_selector.filters.disabled'),
              },
            ]}
            value={detailModelFilter}
            onChange={(e, { value }) =>
              setDetailModelFilter((value || 'all').toString())
            }
          />
          <Form.Input
            className='router-section-input router-search-form-sm'
            icon='search'
            iconPosition='left'
            disabled={detailModelsEditing}
            placeholder={t('channel.edit.model_selector.search_placeholder')}
            value={modelSearchKeyword}
            onChange={(e, { value }) => setModelSearchKeyword(value || '')}
          />
          <Button
            type='button'
            className='router-page-button'
            color='green'
            loading={fetchModelsLoading || !!activeRefreshModelsTask}
            disabled={
              detailModelsEditing ||
              fetchModelsLoading ||
              !!activeRefreshModelsTask ||
              detailModelMutating
            }
            onClick={() => handleFetchModels({ silent: false })}
          >
            {t('channel.edit.buttons.sync_models')}
          </Button>
        </div>
      </div>
      <Form.Field>
        <Message info className='router-section-message'>
          {t('channel.edit.model_selector.enable_hint')}
        </Message>
        <Table
          celled
          stackable
          className='router-detail-table router-channel-detail-model-table'
          compact='very'
        >
          <colgroup>
            {columnWidths.map((width, index) => (
              <col
                key={`channel-detail-model-col-${index}`}
                style={{ width }}
              />
            ))}
          </colgroup>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell
                textAlign='center'
                className='router-create-model-selected-col'
              >
                <div className='router-model-header-checkbox'>
                  <span className='router-model-header-checkbox-label'>
                    {t('channel.edit.model_selector.table.selected')}
                  </span>
                  <Checkbox
                    checked={detailCurrentPageAllSelected}
                    indeterminate={detailCurrentPagePartiallySelected}
                    disabled={
                      detailModelsEditing ||
                      detailModelMutating ||
                      providerCatalogLoading ||
                      detailCurrentPageSelectableCount === 0
                    }
                    onChange={(e, { checked }) =>
                      toggleDetailCurrentPageSelections(!!checked)
                    }
                  />
                </div>
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.name')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.type')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.alias')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.price_unit')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.input_price')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.output_price')}
              </Table.HeaderCell>
              <Table.HeaderCell>{t('channel.table.actions')}</Table.HeaderCell>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {searchedModelConfigs.length === 0 ? (
              <Table.Row>
                <Table.Cell className='router-empty-cell' colSpan={8}>
                  {modelSearchKeyword.trim() !== ''
                    ? t('channel.edit.model_selector.empty_search')
                    : visibleModelConfigs.length > 0
                      ? t('channel.edit.model_selector.empty_filtered')
                      : t('channel.edit.model_selector.empty')}
                </Table.Cell>
              </Table.Row>
            ) : (
              renderedModelConfigs.map((row) => {
                const complexPricingDetails =
                  getComplexPricingDetailsForModel(row);
                const hasComplexInputPricing = complexPricingDetails.some(
                  (detail) =>
                    (detail.price_components || []).some(
                      (component) => Number(component.input_price || 0) > 0,
                    ),
                );
                const hasComplexOutputPricing = complexPricingDetails.some(
                  (detail) =>
                    (detail.price_components || []).some(
                      (component) => Number(component.output_price || 0) > 0,
                    ),
                );
                const rowEditDisabled =
                  detailModelsEditLocked ||
                  detailModelMutating ||
                  detailModelsEditing;
                const rowActionBlocked =
                  !canSelectChannelModel(row) && !row.selected;
                const rowActionDisabled = rowEditDisabled || rowActionBlocked;
                const rowActionDisabledReason = rowActionBlocked
                  ? t(
                      'channel.edit.model_selector.selection_disabled_unassigned',
                    )
                  : '';
                return (
                  <Table.Row key={`${row.upstream_model}-${row.model}`}>
                    {renderModelToggleCells({
                      row,
                      canSelect: canSelectChannelModel(row),
                      selectDisabled:
                        detailModelMutating ||
                        detailModelsEditing ||
                        providerCatalogLoading,
                      inDetailMode: true,
                    })}
                    <Table.Cell
                      title={row.upstream_model}
                      className='router-cell-truncate'
                    >
                      <span className='router-nowrap'>{row.upstream_model}</span>
                      {row.inactive && (
                        <Label basic color='grey' className='router-tag'>
                          {t('channel.edit.model_selector.inactive')}
                        </Label>
                      )}
                    </Table.Cell>
                    <Table.Cell>
                      {t(
                        `channel.model_types.${normalizeChannelModelType(row.type)}`,
                      )}
                    </Table.Cell>
                    <Table.Cell
                      title={row.model}
                      className='router-cell-truncate'
                    >
                      {row.model}
                    </Table.Cell>
                    <Table.Cell>
                      <span className='router-nowrap'>{row.price_unit}</span>
                    </Table.Cell>
                    <Table.Cell>
                      {hasComplexInputPricing ? (
                        <Button
                          type='button'
                          basic
                          className='router-inline-button'
                          onClick={() => openComplexPricingModal(row)}
                        >
                          {t('channel.edit.model_selector.pricing_detail_button')}
                        </Button>
                      ) : (
                        <span className='router-nowrap'>
                          {row.input_price ?? '-'}
                        </span>
                      )}
                    </Table.Cell>
                    <Table.Cell>
                      {hasComplexOutputPricing ? (
                        <Button
                          type='button'
                          basic
                          className='router-inline-button'
                          onClick={() => openComplexPricingModal(row)}
                        >
                          {t('channel.edit.model_selector.pricing_detail_button')}
                        </Button>
                      ) : (
                        <span className='router-nowrap'>
                          {row.output_price ?? '-'}
                        </span>
                      )}
                    </Table.Cell>
                    <Table.Cell collapsing className='router-nowrap'>
                      <div className='router-inline-actions'>
                        <Button
                          type='button'
                          className='router-inline-button'
                          disabled={rowActionDisabled}
                          title={rowActionDisabledReason || undefined}
                          onClick={() => startDetailModelEdit(row.upstream_model)}
                        >
                          {t('common.edit')}
                        </Button>
                      </div>
                    </Table.Cell>
                  </Table.Row>
                );
              })
            )}
          </Table.Body>
        </Table>
        {detailModelTotalPages > 1 && (
          <div className='router-pagination-wrap'>
            <Pagination
              className='router-section-pagination'
              activePage={detailModelPage}
              totalPages={detailModelTotalPages}
              onPageChange={(e, { activePage }) =>
                setDetailModelPage(Number(activePage) || 1)
              }
            />
          </div>
        )}
        {modelsSyncError && (
          <div className='router-error-text router-error-text-top'>
            {modelsSyncError}
          </div>
        )}
      </Form.Field>
    </section>
  );
};

export default ChannelDetailModelsTab;
