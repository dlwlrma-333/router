import React from 'react';
import {
  Button,
  Dropdown,
  Form,
  Label,
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
  detailModelStats,
  searchedModelConfigs,
  visibleModelConfigs,
  renderedModelConfigs,
  getProviderOwnersForModel,
  getSelectedProviderDisplayItems,
  getComplexPricingDetailsForModel,
  openComplexPricingModal,
  detailModelsEditLocked,
  providerCatalogLoading,
  renderModelToggleCells,
  canSelectChannelModel,
  normalizeChannelModelType,
  startDetailModelEdit,
  openAppendProviderModal,
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
        <div className='router-detail-summary-grid'>
          <div className='router-inline-stat-card'>
            <div className='router-inline-stat-value'>
              {detailModelStats.enabled}
            </div>
            <div className='router-inline-stat-hint'>
              {t('channel.edit.model_selector.cards.enabled')}
            </div>
          </div>
          <div className='router-inline-stat-card'>
            <div className='router-inline-stat-value'>
              {detailModelStats.assigned}
            </div>
            <div className='router-inline-stat-hint'>
              {t('channel.edit.model_selector.cards.assigned')}
            </div>
          </div>
          <div className='router-inline-stat-card'>
            <div className='router-inline-stat-value'>
              {detailModelStats.unassigned}
            </div>
            <div className='router-inline-stat-hint'>
              {t('channel.edit.model_selector.cards.unassigned')}
            </div>
          </div>
          <div className='router-inline-stat-card'>
            <div className='router-inline-stat-value'>
              {detailModelStats.inactive}
            </div>
            <div className='router-inline-stat-hint'>
              {t('channel.edit.model_selector.cards.inactive')}
            </div>
          </div>
        </div>
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
                {t('channel.edit.model_selector.table.selected')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.name')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.type')}
              </Table.HeaderCell>
              <Table.HeaderCell>
                {t('channel.edit.model_selector.table.providers')}
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
                <Table.Cell className='router-empty-cell' colSpan={9}>
                  {modelSearchKeyword.trim() !== ''
                    ? t('channel.edit.model_selector.empty_search')
                    : visibleModelConfigs.length > 0
                      ? t('channel.edit.model_selector.empty_filtered')
                      : t('channel.edit.model_selector.empty')}
                </Table.Cell>
              </Table.Row>
            ) : (
              renderedModelConfigs.map((row) => {
                const providerOwners = getProviderOwnersForModel(row);
                const selectedProviderItems = getSelectedProviderDisplayItems(row);
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
                const isUnassigned = providerOwners.length === 0;
                const rowEditDisabled =
                  detailModelsEditLocked ||
                  detailModelMutating ||
                  detailModelsEditing;
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
                    <Table.Cell>
                      <div className='router-create-model-provider-list'>
                        {selectedProviderItems.length > 0 ? (
                          selectedProviderItems.map((provider) => (
                            <Label
                              key={`${row.upstream_model}-${provider.key}`}
                              basic
                              className='router-tag'
                              title={provider.text}
                            >
                              {provider.text}
                            </Label>
                          ))
                        ) : providerOwners.length > 0 ? (
                          providerOwners.map((providerId) => (
                            <Label
                              key={`${row.upstream_model}-${providerId}`}
                              basic
                              className='router-tag'
                            >
                              {providerId}
                            </Label>
                          ))
                        ) : providerCatalogLoading ? (
                          <Label basic className='router-tag'>
                            {t('channel.edit.model_selector.provider_loading')}
                          </Label>
                        ) : (
                          '-'
                        )}
                      </div>
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
                          disabled={rowEditDisabled}
                          onClick={() => startDetailModelEdit(row.upstream_model)}
                        >
                          {t('common.edit')}
                        </Button>
                        {isUnassigned && !providerCatalogLoading ? (
                          <Button
                            type='button'
                            className='router-inline-button'
                            basic
                            disabled={rowEditDisabled}
                            onClick={() => openAppendProviderModal(row)}
                          >
                            {t('channel.edit.model_selector.provider_add')}
                          </Button>
                        ) : null}
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
