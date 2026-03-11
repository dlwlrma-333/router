import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Label,
  Pagination,
  Select,
  Table,
} from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess, timestamp2string } from '../../helpers';

const PAGE_SIZE = 20;

const normalizeTaskStatus = (value) => {
  const normalized = (value || '').toString().trim().toLowerCase();
  switch (normalized) {
    case 'pending':
    case 'running':
    case 'succeeded':
    case 'failed':
    case 'canceled':
      return normalized;
    default:
      return 'pending';
  }
};

const taskStatusColor = (status) => {
  switch (normalizeTaskStatus(status)) {
    case 'running':
      return 'blue';
    case 'succeeded':
      return 'green';
    case 'failed':
      return 'red';
    case 'canceled':
      return 'grey';
    default:
      return 'orange';
  }
};

const taskTypeOptions = (t) => [
  { key: 'all', value: '', text: t('task.filters.type_all') },
  {
    key: 'channel_model_test',
    value: 'channel_model_test',
    text: t('task.types.channel_model_test'),
  },
  {
    key: 'channel_refresh_models',
    value: 'channel_refresh_models',
    text: t('task.types.channel_refresh_models'),
  },
  {
    key: 'channel_refresh_balance',
    value: 'channel_refresh_balance',
    text: t('task.types.channel_refresh_balance'),
  },
];

const taskStatusOptions = (t) => [
  { key: 'all', value: '', text: t('task.filters.status_all') },
  { key: 'pending', value: 'pending', text: t('task.status.pending') },
  { key: 'running', value: 'running', text: t('task.status.running') },
  { key: 'succeeded', value: 'succeeded', text: t('task.status.succeeded') },
  { key: 'failed', value: 'failed', text: t('task.status.failed') },
  { key: 'canceled', value: 'canceled', text: t('task.status.canceled') },
];

const taskQuickStatusOptions = (t) => [
  { key: 'all', value: '', text: t('task.filters.quick_all') },
  { key: 'pending', value: 'pending', text: t('task.status.pending') },
  { key: 'running', value: 'running', text: t('task.status.running') },
  { key: 'failed', value: 'failed', text: t('task.status.failed') },
];

const taskQuickTypeOptions = (t) => [
  { key: 'all', value: '', text: t('task.filters.quick_all') },
  {
    key: 'channel_model_test',
    value: 'channel_model_test',
    text: t('task.types.channel_model_test'),
  },
  {
    key: 'channel_refresh_models',
    value: 'channel_refresh_models',
    text: t('task.types.channel_refresh_models'),
  },
  {
    key: 'channel_refresh_balance',
    value: 'channel_refresh_balance',
    text: t('task.types.channel_refresh_balance'),
  },
];

const Task = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(() => {
    const query = new URLSearchParams(location.search);
    const parsed = Number(query.get('page') || 1);
    return Number.isInteger(parsed) && parsed > 0 ? parsed : 1;
  });
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState(() => {
    const query = new URLSearchParams(location.search);
    return {
      type: (query.get('type') || '').trim(),
      status: (query.get('status') || '').trim(),
      channel_id: (query.get('channel_id') || '').trim(),
    };
  });

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil(total / PAGE_SIZE)),
    [total]
  );

  const loadTasks = useCallback(
    async (targetPage = page) => {
      setLoading(true);
      try {
        const res = await API.get('/api/v1/admin/tasks', {
          params: {
            page: targetPage,
            page_size: PAGE_SIZE,
            type: filters.type,
            status: filters.status,
            channel_id: filters.channel_id.trim(),
          },
        });
        const { success, message, data } = res.data || {};
        if (!success) {
          showError(message || t('task.messages.load_failed'));
          return;
        }
        setItems(Array.isArray(data?.items) ? data.items : []);
        setTotal(Number(data?.total || 0));
        setPage(Number(data?.page || targetPage || 1));
      } catch (error) {
        showError(error?.message || t('task.messages.load_failed'));
      } finally {
        setLoading(false);
      }
    },
    [filters.channel_id, filters.status, filters.type, page, t]
  );

  useEffect(() => {
    loadTasks(1).then();
  }, [loadTasks]);

  useEffect(() => {
    const query = new URLSearchParams();
    if (page > 1) {
      query.set('page', String(page));
    }
    if (filters.type) {
      query.set('type', filters.type);
    }
    if (filters.status) {
      query.set('status', filters.status);
    }
    if (filters.channel_id.trim()) {
      query.set('channel_id', filters.channel_id.trim());
    }
    const nextSearch = query.toString();
    navigate(
      {
        pathname: location.pathname,
        search: nextSearch ? `?${nextSearch}` : '',
      },
      { replace: true }
    );
  }, [filters.channel_id, filters.status, filters.type, location.pathname, navigate, page]);

  useEffect(() => {
    const hasActive = items.some((item) => {
      const status = normalizeTaskStatus(item?.status);
      return status === 'pending' || status === 'running';
    });
    if (!hasActive) {
      return undefined;
    }
    const timer = window.setInterval(() => {
      loadTasks(page).then();
    }, 2500);
    return () => window.clearInterval(timer);
  }, [items, loadTasks, page]);

  const handleRetryTask = async (taskId) => {
    try {
      const res = await API.post(`/api/v1/admin/tasks/${taskId}/retry`);
      const { success, message } = res.data || {};
      if (!success) {
        showError(message || t('task.messages.retry_failed'));
        return;
      }
      showSuccess(t('task.messages.retry_success'));
      loadTasks(page).then();
    } catch (error) {
      showError(error?.message || t('task.messages.retry_failed'));
    }
  };

  const handleCancelTask = async (taskId) => {
    try {
      const res = await API.post(`/api/v1/admin/tasks/${taskId}/cancel`);
      const { success, message } = res.data || {};
      if (!success) {
        showError(message || t('task.messages.cancel_failed'));
        return;
      }
      showSuccess(t('task.messages.cancel_success'));
      loadTasks(page).then();
    } catch (error) {
      showError(error?.message || t('task.messages.cancel_failed'));
    }
  };

  return (
    <div className='dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <div className='router-toolbar router-block-gap-sm'>
            <div className='router-toolbar-start'>
              <Button
                className='router-page-button'
                onClick={() => loadTasks(page)}
                loading={loading}
              >
                {t('task.buttons.refresh')}
              </Button>
              <div className='router-tag-group'>
                {taskQuickStatusOptions(t).map((option) => (
                  <Button
                    key={option.key}
                    type='button'
                    basic={filters.status !== option.value}
                    className='router-inline-button'
                    onClick={() =>
                      setFilters((prev) => ({
                        ...prev,
                        status: option.value,
                      }))
                    }
                  >
                    {option.text}
                  </Button>
                ))}
              </div>
              <div className='router-tag-group'>
                {taskQuickTypeOptions(t).map((option) => (
                  <Button
                    key={option.key}
                    type='button'
                    basic={filters.type !== option.value}
                    className='router-inline-button'
                    onClick={() =>
                      setFilters((prev) => ({
                        ...prev,
                        type: option.value,
                      }))
                    }
                  >
                    {option.text}
                  </Button>
                ))}
              </div>
            </div>
            <div className='router-toolbar-end'>
              <Form>
                <Form.Group widths='equal'>
                  <Form.Input
                    className='router-section-input'
                    placeholder={t('task.filters.channel_id')}
                    value={filters.channel_id}
                    onChange={(e) =>
                      setFilters((prev) => ({
                        ...prev,
                        channel_id: e.target.value,
                      }))
                    }
                  />
                  <Select
                    className='router-section-dropdown'
                    options={taskTypeOptions(t)}
                    value={filters.type}
                    onChange={(e, { value }) =>
                      setFilters((prev) => ({ ...prev, type: value || '' }))
                    }
                  />
                  <Select
                    className='router-section-dropdown'
                    options={taskStatusOptions(t)}
                    value={filters.status}
                    onChange={(e, { value }) =>
                      setFilters((prev) => ({ ...prev, status: value || '' }))
                    }
                  />
                </Form.Group>
              </Form>
            </div>
          </div>

          <Table basic='very' compact className='router-list-table'>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>{t('task.table.type')}</Table.HeaderCell>
                <Table.HeaderCell>{t('task.table.channel')}</Table.HeaderCell>
                <Table.HeaderCell>{t('task.table.model')}</Table.HeaderCell>
                <Table.HeaderCell>{t('task.table.status')}</Table.HeaderCell>
                <Table.HeaderCell>{t('task.table.created_at')}</Table.HeaderCell>
                <Table.HeaderCell>{t('task.table.finished_at')}</Table.HeaderCell>
                <Table.HeaderCell>{t('task.table.message')}</Table.HeaderCell>
                <Table.HeaderCell>{t('task.table.actions')}</Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {items.length === 0 ? (
                <Table.Row>
                  <Table.Cell colSpan='8' className='router-empty-cell'>
                    {loading ? t('common.loading') : t('task.empty')}
                  </Table.Cell>
                </Table.Row>
              ) : (
                items.map((item) => {
                  const status = normalizeTaskStatus(item?.status);
                  const canCancel =
                    status === 'pending' || status === 'running';
                  const canRetry =
                    status === 'failed' || status === 'canceled';
                  const message =
                    item?.error_message || item?.result || item?.payload || '-';
                  return (
                    <Table.Row
                      key={item.id}
                      className='router-row-clickable'
                      onClick={() => navigate(`/admin/task/${item.id}`)}
                    >
                      <Table.Cell>
                        {t(`task.types.${item.type || 'channel_model_test'}`)}
                      </Table.Cell>
                      <Table.Cell>
                        {item.channel_name || item.channel_id || '-'}
                      </Table.Cell>
                      <Table.Cell>{item.model || '-'}</Table.Cell>
                      <Table.Cell>
                        <Label
                          basic
                          color={taskStatusColor(status)}
                          className='router-tag'
                        >
                          {t(`task.status.${status}`)}
                        </Label>
                      </Table.Cell>
                      <Table.Cell>
                        {item.created_at
                          ? timestamp2string(item.created_at)
                          : '-'}
                      </Table.Cell>
                      <Table.Cell>
                        {item.finished_at
                          ? timestamp2string(item.finished_at)
                          : '-'}
                      </Table.Cell>
                      <Table.Cell style={{ maxWidth: 420, wordBreak: 'break-word' }}>
                        {message}
                      </Table.Cell>
                      <Table.Cell collapsing>
                        <div className='router-inline-actions'>
                          <Button
                            type='button'
                            className='router-inline-button'
                            basic
                            disabled={!canRetry}
                            onClick={(e) => {
                              e.stopPropagation();
                              handleRetryTask(item.id);
                            }}
                          >
                            {t('task.buttons.retry')}
                          </Button>
                          <Button
                            type='button'
                            className='router-inline-button'
                            basic
                            disabled={!canCancel}
                            onClick={(e) => {
                              e.stopPropagation();
                              handleCancelTask(item.id);
                            }}
                          >
                            {t('task.buttons.cancel')}
                          </Button>
                        </div>
                      </Table.Cell>
                    </Table.Row>
                  );
                })
              )}
            </Table.Body>
            <Table.Footer>
              <Table.Row>
                <Table.HeaderCell colSpan='8'>
                  <div className='router-toolbar'>
                    <div className='router-toolbar-start'>
                      <span className='router-toolbar-meta'>
                        {t('task.summary', { total })}
                      </span>
                    </div>
                    <Pagination
                      className='router-page-pagination'
                      activePage={page}
                      totalPages={totalPages}
                      onPageChange={(e, { activePage }) => {
                        const nextPage = Number(activePage || 1);
                        setPage(nextPage);
                        loadTasks(nextPage).then();
                      }}
                    />
                  </div>
                </Table.HeaderCell>
              </Table.Row>
            </Table.Footer>
          </Table>
        </Card.Content>
      </Card>
    </div>
  );
};

export default Task;
