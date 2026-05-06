import React from 'react';
import { Button, Checkbox, Dropdown, Form, Message, Modal } from 'semantic-ui-react';

const ChannelEndpointPolicyEditorModal = ({
  t,
  open,
  onClose,
  policyEditorSaving,
  endpointPolicyTemplates,
  selectedPolicyTemplate,
  setSelectedPolicyTemplate,
  applyEndpointPolicyTemplate,
  policyDraft,
  setPolicyDraft,
  saveEndpointPolicy,
}) => {
  return (
    <Modal
      size='large'
      open={open}
      onClose={onClose}
      closeOnDimmerClick={!policyEditorSaving}
    >
      <Modal.Header>
        {t('channel.edit.endpoint_policies.editor.title')}
      </Modal.Header>
      <Modal.Content scrolling>
        <Form>
          <Message info className='router-section-message'>
            {t('channel.edit.endpoint_policies.editor.hint')}
          </Message>
          <Form.Group widths='equal'>
            <Form.Field>
              <label>{t('channel.edit.endpoint_policies.editor.template')}</label>
              <Dropdown
                selection
                clearable
                className='router-modal-dropdown'
                options={endpointPolicyTemplates}
                value={selectedPolicyTemplate}
                placeholder={t(
                  'channel.edit.endpoint_policies.editor.template_placeholder',
                )}
                onChange={(e, { value }) => {
                  const nextValue = (value || '').toString();
                  if (nextValue === '') {
                    setSelectedPolicyTemplate('');
                    return;
                  }
                  applyEndpointPolicyTemplate(nextValue);
                }}
              />
            </Form.Field>
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.Input
              className='router-modal-input'
              label={t('channel.edit.endpoint_policies.table.model')}
              value={policyDraft.model}
              readOnly
            />
            <Form.Input
              className='router-modal-input'
              label={t('channel.edit.endpoint_policies.table.endpoint')}
              value={policyDraft.endpoint}
              readOnly
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.Field>
              <label>{t('channel.edit.endpoint_policies.table.status')}</label>
              <Checkbox
                toggle
                checked={policyDraft.enabled === true}
                onChange={(e, { checked }) =>
                  setPolicyDraft((prev) => ({
                    ...prev,
                    enabled: !!checked,
                  }))
                }
              />
            </Form.Field>
          </Form.Group>
          <Form.TextArea
            className='router-section-textarea router-code-textarea router-code-textarea-sm'
            label={t('channel.edit.endpoint_policies.table.reason')}
            value={policyDraft.reason}
            onChange={(e, { value }) =>
              setPolicyDraft((prev) => ({
                ...prev,
                reason: value || '',
              }))
            }
          />
          <Form.TextArea
            className='router-section-textarea router-code-textarea router-code-textarea-md'
            label={t('channel.edit.endpoint_policies.editor.capabilities')}
            placeholder='{"input_image_url": false}'
            value={policyDraft.capabilities}
            onChange={(e, { value }) =>
              setPolicyDraft((prev) => ({
                ...prev,
                capabilities: value || '',
              }))
            }
          />
          <Form.TextArea
            className='router-section-textarea router-code-textarea router-code-textarea-md'
            label={t('channel.edit.endpoint_policies.editor.request_policy')}
            placeholder='{"actions":[{"type":"drop_fields","fields":["presence_penalty"]}]}'
            value={policyDraft.request_policy}
            onChange={(e, { value }) =>
              setPolicyDraft((prev) => ({
                ...prev,
                request_policy: value || '',
              }))
            }
          />
          <Form.TextArea
            className='router-section-textarea router-code-textarea router-code-textarea-md'
            label={t('channel.edit.endpoint_policies.editor.response_policy')}
            placeholder='{}'
            value={policyDraft.response_policy}
            onChange={(e, { value }) =>
              setPolicyDraft((prev) => ({
                ...prev,
                response_policy: value || '',
              }))
            }
          />
        </Form>
      </Modal.Content>
      <Modal.Actions>
        <Button
          type='button'
          className='router-modal-button'
          onClick={onClose}
          disabled={policyEditorSaving}
        >
          {t('channel.edit.buttons.cancel')}
        </Button>
        <Button
          type='button'
          className='router-modal-button'
          color='blue'
          loading={policyEditorSaving}
          disabled={policyEditorSaving}
          onClick={saveEndpointPolicy}
        >
          {t('channel.edit.buttons.save')}
        </Button>
      </Modal.Actions>
    </Modal>
  );
};

export default ChannelEndpointPolicyEditorModal;
