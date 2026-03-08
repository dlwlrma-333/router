import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Form, Icon, Label, Modal, Table } from 'semantic-ui-react';
import { API, showError, showInfo, showSuccess, timestamp2string } from '../helpers';

const MODE_LIST = 'list';
const MODE_CREATE = 'create';
const MODE_VIEW = 'view';
const MODE_EDIT = 'edit';

const createEmptyForm = () => ({
  id: '',
  name: '',
  description: '',
  billing_ratio: 1,
  sort_order: 0,
});

const sortCatalogRows = (items) =>
  [...items].sort((a, b) => {
    const aOrder = Number(a.sort_order || 0);
    const bOrder = Number(b.sort_order || 0);
    if (aOrder !== bOrder) {
      return aOrder - bOrder;
    }
    return (a.id || '').localeCompare(b.id || '');
  });

const buildFormFromRow = (row) => ({
  id: row?.id || '',
  name: row?.name || '',
  description: row?.description || '',
  billing_ratio: Number(row?.billing_ratio ?? 1),
  sort_order: Number(row?.sort_order || 0),
});

const toChannelOptions = (items) =>
  (Array.isArray(items) ? items : []).map((item) => ({
    key: item.id,
    text: `${item.name || item.id} (${item.id})`,
    value: item.id,
  }));

const toBoundChannelIDs = (items) =>
  (Array.isArray(items) ? items : [])
    .filter((item) => !!item.bound)
    .map((item) => item.id);

const toBoundChannelRows = (items) =>
  (Array.isArray(items) ? items : []).filter((item) => !!item.bound);

const actionBarStyle = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'flex-start',
  gap: '8px',
  flexWrap: 'wrap',
  marginBottom: 12,
};

const GroupsManager = () => {
  const { t } = useTranslation();
  const [mode, setMode] = useState(MODE_LIST);
  const [rows, setRows] = useState([]);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');

  const [activeGroup, setActiveGroup] = useState(null);
  const [form, setForm] = useState(createEmptyForm());
  const [formChannelOptions, setFormChannelOptions] = useState([]);
  const [formChannelIDs, setFormChannelIDs] = useState([]);
  const [formChannelLoading, setFormChannelLoading] = useState(false);

  const [detailChannelRows, setDetailChannelRows] = useState([]);
  const [detailChannelLoading, setDetailChannelLoading] = useState(false);

  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);

  const loadCatalog = useCallback(async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/v1/admin/group/catalog');
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('group_manage.messages.load_failed'));
        return;
      }
      setRows(sortCatalogRows(Array.isArray(data) ? data : []));
    } catch (error) {
      showError(error);
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    loadCatalog().then();
  }, [loadCatalog]);

  const visibleRows = useMemo(() => {
    const keyword = typeof searchKeyword === 'string' ? searchKeyword.trim().toLowerCase() : '';
    if (!keyword) {
      return rows;
    }
    return rows.filter((row) => {
      const haystacks = [row.id, row.name, row.description];
      return haystacks.some((item) =>
        typeof item === 'string' ? item.toLowerCase().includes(keyword) : false
      );
    });
  }, [rows, searchKeyword]);

  const resetFormState = () => {
    setForm(createEmptyForm());
    setFormChannelOptions([]);
    setFormChannelIDs([]);
    setFormChannelLoading(false);
  };

  const resetDetailState = () => {
    setDetailChannelRows([]);
    setDetailChannelLoading(false);
  };

  const clearDeleteState = () => {
    setDeleteOpen(false);
    setDeleteTarget(null);
  };

  const closeDeleteModal = () => {
    if (submitting) return;
    clearDeleteState();
  };

  const fetchCreateChannelOptions = useCallback(async () => {
    setFormChannelLoading(true);
    try {
      const res = await API.get('/api/v1/admin/group/channel-options');
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('group_manage.messages.bind_load_failed'));
        return;
      }
      setFormChannelOptions(toChannelOptions(data));
      setFormChannelIDs([]);
    } catch (error) {
      showError(error);
    } finally {
      setFormChannelLoading(false);
    }
  }, [t]);

  const fetchGroupChannels = useCallback(async (groupID) => {
    const encodedID = encodeURIComponent(groupID || '');
    const res = await API.get(`/api/v1/admin/group/${encodedID}/channels`);
    const { success, message, data } = res.data || {};
    if (!success) {
      throw new Error(message || t('group_manage.messages.bind_load_failed'));
    }
    return Array.isArray(data) ? data : [];
  }, [t]);

  const loadViewChannelRows = useCallback(async (groupID) => {
    setDetailChannelLoading(true);
    try {
      const rows = await fetchGroupChannels(groupID);
      setDetailChannelRows(toBoundChannelRows(rows));
    } catch (error) {
      showError(error);
    } finally {
      setDetailChannelLoading(false);
    }
  }, [fetchGroupChannels]);

  const loadEditChannelRows = useCallback(async (groupID) => {
    setFormChannelLoading(true);
    try {
      const rows = await fetchGroupChannels(groupID);
      setFormChannelOptions(toChannelOptions(rows));
      setFormChannelIDs(toBoundChannelIDs(rows));
    } catch (error) {
      showError(error);
    } finally {
      setFormChannelLoading(false);
    }
  }, [fetchGroupChannels]);

  const resetToList = () => {
    setMode(MODE_LIST);
    setActiveGroup(null);
    resetFormState();
    resetDetailState();
  };

  const backToList = () => {
    if (submitting) return;
    resetToList();
  };

  const openCreatePanel = () => {
    if (submitting) return;
    setMode(MODE_CREATE);
    setActiveGroup(null);
    resetDetailState();
    resetFormState();
    fetchCreateChannelOptions().then();
  };

  const openViewPanel = (row) => {
    if (!row || submitting) return;
    setMode(MODE_VIEW);
    setActiveGroup(row);
    resetFormState();
    resetDetailState();
    loadViewChannelRows(row.id || '').then();
  };

  const openEditPanel = (row = activeGroup) => {
    if (!row || submitting) return;
    setMode(MODE_EDIT);
    setActiveGroup(row);
    setForm(buildFormFromRow(row));
    setFormChannelOptions([]);
    setFormChannelIDs([]);
    loadEditChannelRows(row.id || '').then();
  };

  const submitCreate = async () => {
    const id = (form.id || '').trim();
    if (id === '') {
      showInfo(t('group_manage.messages.id_required'));
      return;
    }
    const billingRatio = Number(form.billing_ratio ?? 1);
    if (!Number.isFinite(billingRatio) || billingRatio < 0) {
      showInfo(t('group_manage.messages.billing_ratio_invalid'));
      return;
    }
    setSubmitting(true);
    try {
      const res = await API.post('/api/v1/admin/group/', {
        id,
        name: (form.name || '').trim(),
        description: (form.description || '').trim(),
        billing_ratio: billingRatio,
        channel_ids: formChannelIDs,
      });
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('group_manage.messages.create_failed'));
        return;
      }
      setRows((prev) => sortCatalogRows([...prev, data]));
      showSuccess(t('group_manage.messages.create_success'));
      resetToList();
    } catch (error) {
      showError(error);
    } finally {
      setSubmitting(false);
    }
  };

  const submitEdit = async () => {
    const id = (form.id || '').trim();
    if (id === '') {
      showInfo(t('group_manage.messages.id_required'));
      return;
    }
    const billingRatio = Number(form.billing_ratio ?? 1);
    if (!Number.isFinite(billingRatio) || billingRatio < 0) {
      showInfo(t('group_manage.messages.billing_ratio_invalid'));
      return;
    }
    setSubmitting(true);
    try {
      const res = await API.put('/api/v1/admin/group/', {
        id,
        name: (form.name || '').trim(),
        description: (form.description || '').trim(),
        billing_ratio: billingRatio,
        sort_order: Number(form.sort_order || 0),
        channel_ids: formChannelIDs,
      });
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('group_manage.messages.update_failed'));
        return;
      }
      setRows((prev) =>
        sortCatalogRows(prev.map((row) => (row.id === data.id ? data : row)))
      );
      setActiveGroup(data);
      showSuccess(t('group_manage.messages.update_success'));
      setMode(MODE_VIEW);
      resetFormState();
      loadViewChannelRows(data.id || '').then();
    } catch (error) {
      showError(error);
    } finally {
      setSubmitting(false);
    }
  };

  const toggleEnabled = async (row) => {
    if (!row || submitting) return;
    setSubmitting(true);
    try {
      const res = await API.put('/api/v1/admin/group/', {
        id: row.id,
        name: row.name || '',
        description: row.description || '',
        billing_ratio: Number(row.billing_ratio ?? 1),
        sort_order: Number(row.sort_order || 0),
        enabled: !row.enabled,
      });
      const { success, message, data } = res.data || {};
      if (!success) {
        showError(message || t('group_manage.messages.update_failed'));
        return;
      }
      setRows((prev) =>
        sortCatalogRows(prev.map((item) => (item.id === data.id ? data : item)))
      );
      if (activeGroup?.id === data.id) {
        setActiveGroup(data);
      }
      showSuccess(t('group_manage.messages.update_success'));
    } catch (error) {
      showError(error);
    } finally {
      setSubmitting(false);
    }
  };

  const openDeleteModal = (row) => {
    if (!row || submitting) return;
    setDeleteTarget(row);
    setDeleteOpen(true);
  };

  const submitDelete = async () => {
    if (!deleteTarget || submitting) return;
    setSubmitting(true);
    try {
      const encodedID = encodeURIComponent(deleteTarget.id || '');
      const res = await API.delete(`/api/v1/admin/group/${encodedID}`);
      const { success, message } = res.data || {};
      if (!success) {
        showError(message || t('group_manage.messages.delete_failed'));
        return;
      }
      setRows((prev) => prev.filter((row) => row.id !== deleteTarget.id));
      showSuccess(t('group_manage.messages.delete_success'));
      clearDeleteState();
      if (activeGroup?.id === deleteTarget.id) {
        resetToList();
      }
    } catch (error) {
      showError(error);
    } finally {
      setSubmitting(false);
    }
  };

  const renderGroupStatus = (enabled) =>
    enabled ? (
      <Label basic color='green'>
        {t('group_manage.status.enabled')}
      </Label>
    ) : (
      <Label basic color='grey'>
        {t('group_manage.status.disabled')}
      </Label>
    );

  const renderChannelStatus = (status) => {
    const normalized = Number(status || 0);
    if (normalized === 1) {
      return (
        <Label basic color='green'>
          {t('channel.table.status_enabled')}
        </Label>
      );
    }
    if (normalized === 4) {
      return (
        <Label basic color='blue'>
          {t('channel.table.status_creating')}
        </Label>
      );
    }
    return (
      <Label basic color='grey'>
        {t('channel.table.status_disabled')}
      </Label>
    );
  };

  const renderList = () => (
    <>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '12px',
          flexWrap: 'wrap',
          marginBottom: '12px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
          <Button
            type='button'
            size='tiny'
            disabled={submitting || loading}
            onClick={openCreatePanel}
          >
            {t('group_manage.buttons.add')}
          </Button>
          <Button
            type='button'
            size='tiny'
            disabled={submitting}
            loading={loading}
            onClick={loadCatalog}
          >
            {t('group_manage.buttons.refresh')}
          </Button>
        </div>
        <Form style={{ width: '320px', maxWidth: '100%' }}>
          <Form.Input
            icon='search'
            iconPosition='left'
            placeholder={t('group_manage.search')}
            value={searchKeyword}
            onChange={(e, { value }) => setSearchKeyword(value || '')}
          />
        </Form>
      </div>

      <Table basic='very' compact size='small' className='router-hover-table'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>{t('group_manage.table.id')}</Table.HeaderCell>
            <Table.HeaderCell>{t('group_manage.table.name')}</Table.HeaderCell>
            <Table.HeaderCell>{t('group_manage.table.description')}</Table.HeaderCell>
            <Table.HeaderCell>{t('group_manage.table.billing_ratio')}</Table.HeaderCell>
            <Table.HeaderCell>{t('group_manage.table.status')}</Table.HeaderCell>
            <Table.HeaderCell>{t('group_manage.table.sort_order')}</Table.HeaderCell>
            <Table.HeaderCell>{t('group_manage.table.updated_at')}</Table.HeaderCell>
            <Table.HeaderCell style={{ width: '220px' }}>
              {t('group_manage.table.actions')}
            </Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {visibleRows.length === 0 ? (
            <Table.Row>
              <Table.Cell colSpan={8} textAlign='center'>
                {loading
                  ? t('group_manage.messages.loading')
                  : t('group_manage.messages.empty')}
              </Table.Cell>
            </Table.Row>
          ) : (
            visibleRows.map((row) => (
              <Table.Row
                key={row.id}
                onClick={() => openViewPanel(row)}
                style={{ cursor: submitting || loading ? 'default' : 'pointer' }}
              >
                <Table.Cell>{row.id}</Table.Cell>
                <Table.Cell>{row.name || '-'}</Table.Cell>
                <Table.Cell>{row.description || '-'}</Table.Cell>
                <Table.Cell>{Number(row.billing_ratio ?? 1).toFixed(2)}</Table.Cell>
                <Table.Cell>{renderGroupStatus(row.enabled)}</Table.Cell>
                <Table.Cell>{row.sort_order || 0}</Table.Cell>
                <Table.Cell>{row.updated_at ? timestamp2string(row.updated_at) : '-'}</Table.Cell>
                <Table.Cell>
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: '6px',
                      flexWrap: 'wrap',
                    }}
                  >
                    <Button
                      size='tiny'
                      color={row.enabled ? 'orange' : 'green'}
                      disabled={submitting || loading}
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleEnabled(row);
                      }}
                    >
                      {row.enabled
                        ? t('group_manage.buttons.disable')
                        : t('group_manage.buttons.enable')}
                    </Button>
                    <Button
                      size='tiny'
                      negative
                      disabled={submitting || loading}
                      onClick={(e) => {
                        e.stopPropagation();
                        openDeleteModal(row);
                      }}
                    >
                      {t('group_manage.buttons.delete')}
                    </Button>
                  </div>
                </Table.Cell>
              </Table.Row>
            ))
          )}
        </Table.Body>
      </Table>
    </>
  );

  const renderBoundChannelsTable = (items, loadingState) => (
    <div style={{ marginTop: 12 }}>
      <div style={{ fontWeight: 600, marginBottom: 8 }}>
        {t('group_manage.detail.bound_channels')}
      </div>
      <Table compact size='small' celled>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>{t('channel.table.id')}</Table.HeaderCell>
            <Table.HeaderCell>{t('channel.table.name')}</Table.HeaderCell>
            <Table.HeaderCell>{t('channel.table.type')}</Table.HeaderCell>
            <Table.HeaderCell>{t('channel.table.status')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {loadingState ? (
            <Table.Row>
              <Table.Cell colSpan={4} textAlign='center'>
                {t('group_manage.messages.loading')}
              </Table.Cell>
            </Table.Row>
          ) : items.length === 0 ? (
            <Table.Row>
              <Table.Cell colSpan={4} textAlign='center'>
                {t('group_manage.detail.empty_channels')}
              </Table.Cell>
            </Table.Row>
          ) : (
            items.map((item) => (
              <Table.Row key={item.id}>
                <Table.Cell>{item.id}</Table.Cell>
                <Table.Cell>{item.name || '-'}</Table.Cell>
                <Table.Cell>{item.protocol || '-'}</Table.Cell>
                <Table.Cell>{renderChannelStatus(item.status)}</Table.Cell>
              </Table.Row>
            ))
          )}
        </Table.Body>
      </Table>
    </div>
  );

  const renderView = () => {
    if (!activeGroup) return null;
    return (
      <div>
        <div style={actionBarStyle}>
          <Button type='button' onClick={backToList} disabled={submitting}>
            <Icon name='undo' />
            {t('group_manage.buttons.back')}
          </Button>
          <Button type='button' color='blue' disabled={submitting} onClick={() => openEditPanel()}>
            <Icon name='edit' />
            {t('group_manage.buttons.edit')}
          </Button>
        </div>
        <Form>
          <Form.Group widths='equal'>
            <Form.Input
              label={t('group_manage.form.id')}
              value={activeGroup.id || ''}
              readOnly
            />
            <Form.Input
              label={t('group_manage.form.name')}
              value={activeGroup.name || ''}
              readOnly
            />
          </Form.Group>
          <Form.TextArea
            label={t('group_manage.form.description')}
            value={activeGroup.description || ''}
            readOnly
          />
          <Form.Group widths='equal'>
            <Form.Input
              label={t('group_manage.form.billing_ratio')}
              value={Number(activeGroup.billing_ratio ?? 1).toFixed(2)}
              readOnly
            />
            <Form.Input
              label={t('group_manage.table.status')}
              value={
                activeGroup.enabled
                  ? t('group_manage.status.enabled')
                  : t('group_manage.status.disabled')
              }
              readOnly
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.Input
              label={t('group_manage.form.sort_order')}
              value={activeGroup.sort_order || 0}
              readOnly
            />
            <Form.Input
              label={t('group_manage.table.updated_at')}
              value={activeGroup.updated_at ? timestamp2string(activeGroup.updated_at) : '-'}
              readOnly
            />
          </Form.Group>
        </Form>
        {renderBoundChannelsTable(detailChannelRows, detailChannelLoading)}
      </div>
    );
  };

  const renderEdit = () => (
    <div>
      <div style={actionBarStyle}>
        <Button type='button' onClick={() => setMode(MODE_VIEW)} disabled={submitting}>
          {t('group_manage.buttons.cancel')}
        </Button>
        <Button type='button' color='blue' loading={submitting} disabled={submitting} onClick={submitEdit}>
          <Icon name='check' />
          {t('group_manage.buttons.confirm')}
        </Button>
      </div>
      <Form>
        <Form.Group widths='equal'>
          <Form.Input
            label={t('group_manage.form.id')}
            value={form.id}
            readOnly
          />
          <Form.Input
            label={t('group_manage.form.name')}
            placeholder={t('group_manage.form.name_placeholder')}
            value={form.name}
            onChange={(e, { value }) =>
              setForm((prev) => ({ ...prev, name: value || '' }))
            }
          />
        </Form.Group>
        <Form.TextArea
          label={t('group_manage.form.description')}
          placeholder={t('group_manage.form.description_placeholder')}
          value={form.description}
          onChange={(e) =>
            setForm((prev) => ({
              ...prev,
              description: e.target.value,
            }))
          }
        />
        <Form.Group widths='equal'>
          <Form.Input
            type='number'
            min='0'
            step='0.01'
            label={t('group_manage.form.billing_ratio')}
            placeholder={t('group_manage.form.billing_ratio_placeholder')}
            value={form.billing_ratio}
            onChange={(e) =>
              setForm((prev) => ({
                ...prev,
                billing_ratio: e.target.value,
              }))
            }
          />
          <Form.Input
            type='number'
            label={t('group_manage.form.sort_order')}
            value={form.sort_order}
            onChange={(e) =>
              setForm((prev) => ({
                ...prev,
                sort_order: Number(e.target.value || 0),
              }))
            }
          />
        </Form.Group>
        <Form.Dropdown
          fluid
          multiple
          search
          selection
          loading={formChannelLoading}
          disabled={formChannelLoading || submitting}
          label={t('group_manage.form.channels')}
          placeholder={t('group_manage.form.channels_placeholder')}
          options={formChannelOptions}
          value={formChannelIDs}
          onChange={(e, { value }) =>
            setFormChannelIDs(Array.isArray(value) ? value : [])
          }
        />
      </Form>
    </div>
  );

  const renderCreate = () => (
    <div>
      <div style={actionBarStyle}>
        <Button type='button' onClick={backToList} disabled={submitting}>
          {t('group_manage.buttons.cancel')}
        </Button>
        <Button type='button' color='blue' loading={submitting} disabled={submitting} onClick={submitCreate}>
          <Icon name='check' />
          {t('group_manage.buttons.confirm')}
        </Button>
      </div>
      <Form>
        <Form.Group widths='equal'>
          <Form.Input
            required
            label={t('group_manage.form.id')}
            placeholder={t('group_manage.form.id_placeholder')}
            value={form.id}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, id: e.target.value }))
            }
          />
          <Form.Input
            label={t('group_manage.form.name')}
            placeholder={t('group_manage.form.name_placeholder')}
            value={form.name}
            onChange={(e) =>
              setForm((prev) => ({ ...prev, name: e.target.value }))
            }
          />
        </Form.Group>
        <Form.TextArea
          label={t('group_manage.form.description')}
          placeholder={t('group_manage.form.description_placeholder')}
          value={form.description}
          onChange={(e) =>
            setForm((prev) => ({
              ...prev,
              description: e.target.value,
            }))
          }
        />
        <Form.Group widths='equal'>
          <Form.Input
            type='number'
            min='0'
            step='0.01'
            label={t('group_manage.form.billing_ratio')}
            placeholder={t('group_manage.form.billing_ratio_placeholder')}
            value={form.billing_ratio}
            onChange={(e) =>
              setForm((prev) => ({
                ...prev,
                billing_ratio: e.target.value,
              }))
            }
          />
          <Form.Dropdown
            fluid
            multiple
            search
            selection
            loading={formChannelLoading}
            disabled={formChannelLoading || submitting}
            label={t('group_manage.form.channels')}
            placeholder={t('group_manage.form.channels_placeholder')}
            options={formChannelOptions}
            value={formChannelIDs}
            onChange={(e, { value }) =>
              setFormChannelIDs(Array.isArray(value) ? value : [])
            }
          />
        </Form.Group>
      </Form>
    </div>
  );

  return (
    <>
      {mode === MODE_CREATE
        ? renderCreate()
        : mode === MODE_EDIT
          ? renderEdit()
          : mode === MODE_VIEW
            ? renderView()
            : renderList()}

      <Modal open={deleteOpen} onClose={closeDeleteModal} size='tiny'>
        <Modal.Header>{t('group_manage.modal.delete_title')}</Modal.Header>
        <Modal.Content>
          {t('group_manage.modal.delete_confirm', {
            id: deleteTarget?.id || '',
          })}
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={closeDeleteModal} disabled={submitting}>
            {t('group_manage.buttons.cancel')}
          </Button>
          <Button negative onClick={submitDelete} loading={submitting}>
            {t('group_manage.buttons.confirm')}
          </Button>
        </Modal.Actions>
      </Modal>
    </>
  );
};

export default GroupsManager;
