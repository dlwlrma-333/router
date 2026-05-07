import React from 'react';
import { Button, Form } from 'semantic-ui-react';

const ChannelDetailOverviewTab = ({
  t,
  inputs,
  currentProtocolOption,
  channelProtocolOptions,
  detailBasicEditing,
  detailBasicSaving,
  detailBasicEditLocked,
  detailBasicReadonly,
  channelIdentifierMaxLength,
  handleInputChange,
  cancelDetailBasicEdit,
  saveDetailBasicInfo,
  setDetailBasicEditing,
  basicConnectionFields,
  protocolSelectionHintContent,
  protocolSpecificFields,
  timestamp2string,
}) => {
  return (
    <>
      <section className='router-entity-detail-section'>
        <div className='router-entity-detail-section-header'>
          <div className='router-toolbar-start'>
            <span className='router-entity-detail-section-title'>
              {t('channel.edit.detail_basic_title')}
            </span>
          </div>
          <div className='router-toolbar-end'>
            {detailBasicEditing ? (
              <>
                <Button
                  type='button'
                  className='router-page-button'
                  onClick={cancelDetailBasicEdit}
                  disabled={detailBasicSaving}
                >
                  {t('channel.edit.buttons.cancel')}
                </Button>
                <Button
                  type='button'
                  className='router-page-button'
                  color='blue'
                  loading={detailBasicSaving}
                  disabled={detailBasicSaving}
                  onClick={saveDetailBasicInfo}
                >
                  {t('channel.edit.buttons.save')}
                </Button>
              </>
            ) : (
              <Button
                type='button'
                className='router-page-button'
                color='blue'
                disabled={detailBasicEditLocked}
                onClick={() => setDetailBasicEditing(true)}
              >
                {t('common.edit')}
              </Button>
            )}
          </div>
        </div>
        <Form.Group widths='equal'>
          <Form.Input
            className='router-section-input'
            label={t('channel.edit.id')}
            value={inputs.id || '-'}
            readOnly
          />
          <Form.Input
            className='router-section-input'
            label={t('channel.edit.identifier')}
            name='name'
            placeholder={t('channel.edit.identifier_placeholder')}
            onChange={handleInputChange}
            value={inputs.name}
            required
            maxLength={channelIdentifierMaxLength}
            readOnly={detailBasicReadonly}
          />
          <Form.Field>
            {detailBasicReadonly ? (
              <Form.Input
                className='router-section-input'
                label={t('channel.edit.type')}
                value={currentProtocolOption?.text || inputs.protocol || '-'}
                readOnly
              />
            ) : (
              <Form.Select
                className='router-section-dropdown'
                label={t('channel.edit.type')}
                name='protocol'
                required
                search
                options={channelProtocolOptions}
                value={inputs.protocol}
                onChange={handleInputChange}
              />
            )}
          </Form.Field>
        </Form.Group>
        {protocolSelectionHintContent}
        {basicConnectionFields}
        {protocolSpecificFields}
        <Form.Group widths='equal'>
          <Form.Input
            className='router-section-input'
            label={t('channel.edit.created_time')}
            value={
              inputs.created_time ? timestamp2string(inputs.created_time) : '-'
            }
            readOnly
          />
          <Form.Input
            className='router-section-input'
            label={t('channel.edit.updated_at')}
            value={inputs.updated_at ? timestamp2string(inputs.updated_at) : '-'}
            readOnly
          />
        </Form.Group>
      </section>
    </>
  );
};

export default ChannelDetailOverviewTab;
