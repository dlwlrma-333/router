import React from 'react';
import { Button, Dropdown, Form, Modal } from 'semantic-ui-react';

const ChannelAppendProviderModal = ({
  t,
  open,
  onClose,
  appendingProviderModel,
  filterProviderOptionsByQuery,
  providerOptions,
  appendProviderForm,
  setAppendProviderForm,
  channelModelTypeOptions,
  normalizeChannelModelType,
  handleAppendModelToProvider,
}) => {
  return (
    <Modal
      size='tiny'
      open={open}
      onClose={onClose}
      closeOnDimmerClick={!appendingProviderModel}
    >
      <Modal.Header>
        {t('channel.edit.model_selector.append_dialog.title')}
      </Modal.Header>
      <Modal.Content>
        <Form>
          <Form.Field>
            <label>{t('channel.edit.model_selector.append_dialog.provider')}</label>
            <Dropdown
              selection
              search={filterProviderOptionsByQuery}
              className='router-modal-dropdown'
              placeholder={t(
                'channel.edit.model_selector.append_dialog.provider_placeholder',
              )}
              options={providerOptions}
              value={appendProviderForm.provider}
              noResultsMessage={t('common.no_data')}
              onChange={(e, { value }) =>
                setAppendProviderForm((prev) => ({
                  ...prev,
                  provider: (value || '').toString(),
                }))
              }
            />
          </Form.Field>
          <Form.Input
            className='router-modal-input'
            label={t('channel.edit.model_selector.append_dialog.model')}
            value={appendProviderForm.model}
            onChange={(e, { value }) =>
              setAppendProviderForm((prev) => ({
                ...prev,
                model: value || '',
              }))
            }
          />
          <Form.Select
            className='router-modal-dropdown'
            label={t('channel.edit.model_selector.append_dialog.type')}
            options={channelModelTypeOptions}
            value={appendProviderForm.type}
            onChange={(e, { value }) =>
              setAppendProviderForm((prev) => ({
                ...prev,
                type: normalizeChannelModelType(value),
              }))
            }
          />
        </Form>
      </Modal.Content>
      <Modal.Actions>
        <Button type='button' className='router-modal-button' onClick={onClose}>
          {t('channel.edit.model_selector.append_dialog.cancel')}
        </Button>
        <Button
          type='button'
          className='router-modal-button'
          color='blue'
          loading={appendingProviderModel}
          disabled={appendingProviderModel}
          onClick={handleAppendModelToProvider}
        >
          {t('channel.edit.model_selector.append_dialog.confirm')}
        </Button>
      </Modal.Actions>
    </Modal>
  );
};

export default ChannelAppendProviderModal;
