import React from 'react';
import { Button, Checkbox, Dropdown, Form, Modal } from 'semantic-ui-react';

const ChannelModelEditorModal = ({
  t,
  open,
  onClose,
  detailModelMutating,
  detailEditingModelRow,
  normalizeChannelModelType,
  updateModelConfigField,
  providerCatalogLoading,
  getProviderSelectOptionsForModel,
  resolvePreferredProviderForModel,
  openAppendProviderModal,
  canSelectChannelModel,
  toggleModelSelection,
  getComplexPricingDetailsForModel,
  openComplexPricingModal,
  saveDetailModelsConfig,
}) => {
  return (
    <Modal
      size='small'
      open={open}
      onClose={onClose}
      closeOnDimmerClick={!detailModelMutating}
      closeOnEscape={!detailModelMutating}
      className='router-channel-model-editor-modal'
    >
      <Modal.Header>
        {`${t('common.edit')} · ${detailEditingModelRow?.upstream_model || '-'}`}
      </Modal.Header>
      <Modal.Content>
        {detailEditingModelRow ? (
          <Form className='router-channel-model-editor-form'>
            <div className='router-channel-model-editor-card'>
              <div className='router-channel-model-editor-section-title'>
                {t('channel.edit.model_selector.editor.info_title')}
              </div>
              <Form.Group widths='equal'>
                <Form.Input
                  className='router-modal-input'
                  label={t('channel.edit.model_selector.table.name')}
                  value={detailEditingModelRow.upstream_model || '-'}
                  readOnly
                />
                <Form.Input
                  className='router-modal-input'
                  label={t('channel.edit.model_selector.table.type')}
                  value={t(
                    `channel.model_types.${normalizeChannelModelType(detailEditingModelRow.type)}`,
                  )}
                  readOnly
                />
              </Form.Group>
              <Form.Group widths='equal'>
                <Form.Input
                  className='router-modal-input'
                  label={t('channel.edit.model_selector.table.alias')}
                  value={detailEditingModelRow.model || ''}
                  onChange={(e, { value }) =>
                    updateModelConfigField(
                      detailEditingModelRow.upstream_model,
                      'model',
                      value || detailEditingModelRow.upstream_model,
                    )
                  }
                />
                <Form.Input
                  className='router-modal-input'
                  label={t('channel.edit.model_selector.table.price_unit')}
                  value={detailEditingModelRow.price_unit || '-'}
                  readOnly
                />
              </Form.Group>
              <Form.Field>
                <label>{t('channel.edit.model_selector.table.providers')}</label>
                <div className='router-channel-model-editor-provider-row'>
                  <Dropdown
                    selection
                    fluid
                    className='router-modal-dropdown'
                    placeholder={t(
                      'channel.edit.model_selector.editor.provider_placeholder',
                    )}
                    options={getProviderSelectOptionsForModel(
                      detailEditingModelRow,
                    )}
                    value={resolvePreferredProviderForModel(
                      detailEditingModelRow,
                    )}
                    disabled={
                      providerCatalogLoading ||
                      getProviderSelectOptionsForModel(detailEditingModelRow)
                        .length === 0
                    }
                    onChange={(e, { value }) =>
                      updateModelConfigField(
                        detailEditingModelRow.upstream_model,
                        'provider',
                        value || '',
                      )
                    }
                  />
                  {getProviderSelectOptionsForModel(detailEditingModelRow)
                    .length === 0 ? (
                    <>
                      <span className='router-text-meta'>
                        {t('channel.edit.model_selector.editor.provider_empty')}
                      </span>
                      <Button
                        type='button'
                        className='router-inline-button'
                        basic
                        onClick={() => openAppendProviderModal(detailEditingModelRow)}
                      >
                        {t('channel.edit.model_selector.provider_add')}
                      </Button>
                    </>
                  ) : null}
                </div>
              </Form.Field>
            </div>

            <div className='router-channel-model-editor-card'>
              <div className='router-channel-model-editor-section-title'>
                {t('channel.edit.model_selector.editor.status_title')}
              </div>
              <div className='router-channel-model-editor-toggle-row'>
                <div className='router-channel-model-editor-toggle-copy'>
                  <div className='router-channel-model-editor-toggle-label'>
                    {t('channel.edit.model_selector.table.selected')}
                  </div>
                  <div className='router-channel-model-editor-toggle-hint'>
                    {t('channel.edit.model_selector.editor.status_hint')}
                  </div>
                </div>
                <Checkbox
                  toggle
                  checked={!!detailEditingModelRow.selected}
                  disabled={
                    detailModelMutating ||
                    providerCatalogLoading ||
                    (!canSelectChannelModel(detailEditingModelRow) &&
                      !detailEditingModelRow.selected)
                  }
                  onChange={(e, { checked }) =>
                    toggleModelSelection(
                      detailEditingModelRow.upstream_model,
                      checked,
                    )
                  }
                />
              </div>
            </div>

            <div className='router-channel-model-editor-card'>
              <div className='router-channel-model-editor-section-title'>
                {t('channel.edit.model_selector.editor.pricing_title')}
              </div>
              <Form.Group widths='equal'>
                <Form.Field>
                  <label>{t('channel.edit.model_selector.table.input_price')}</label>
                  {getComplexPricingDetailsForModel(detailEditingModelRow).some(
                    (detail) =>
                      (detail.price_components || []).some(
                        (component) => Number(component.input_price || 0) > 0,
                      ),
                  ) ? (
                    <Button
                      type='button'
                      basic
                      className='router-inline-button'
                      onClick={() => openComplexPricingModal(detailEditingModelRow)}
                    >
                      {t('channel.edit.model_selector.pricing_detail_button')}
                    </Button>
                  ) : (
                    <Form.Input
                      className='router-modal-input'
                      type='number'
                      min='0'
                      step='0.01'
                      placeholder='-'
                      value={detailEditingModelRow.input_price ?? ''}
                      onChange={(e, { value }) =>
                        updateModelConfigField(
                          detailEditingModelRow.upstream_model,
                          'input_price',
                          value,
                        )
                      }
                    />
                  )}
                </Form.Field>
                <Form.Field>
                  <label>{t('channel.edit.model_selector.table.output_price')}</label>
                  {getComplexPricingDetailsForModel(detailEditingModelRow).some(
                    (detail) =>
                      (detail.price_components || []).some(
                        (component) => Number(component.output_price || 0) > 0,
                      ),
                  ) ? (
                    <Button
                      type='button'
                      basic
                      className='router-inline-button'
                      onClick={() => openComplexPricingModal(detailEditingModelRow)}
                    >
                      {t('channel.edit.model_selector.pricing_detail_button')}
                    </Button>
                  ) : (
                    <Form.Input
                      className='router-modal-input'
                      type='number'
                      min='0'
                      step='0.01'
                      placeholder='-'
                      value={detailEditingModelRow.output_price ?? ''}
                      onChange={(e, { value }) =>
                        updateModelConfigField(
                          detailEditingModelRow.upstream_model,
                          'output_price',
                          value,
                        )
                      }
                    />
                  )}
                </Form.Field>
              </Form.Group>
            </div>
          </Form>
        ) : null}
      </Modal.Content>
      <Modal.Actions>
        <Button
          type='button'
          className='router-modal-button'
          onClick={onClose}
          disabled={detailModelMutating}
        >
          {t('channel.edit.buttons.cancel')}
        </Button>
        <Button
          type='button'
          className='router-modal-button'
          color='blue'
          loading={detailModelMutating}
          disabled={detailModelMutating}
          onClick={saveDetailModelsConfig}
        >
          {t('channel.edit.buttons.save')}
        </Button>
      </Modal.Actions>
    </Modal>
  );
};

export default ChannelModelEditorModal;
