import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Card, Dropdown, Label, Table } from 'semantic-ui-react';
import { API } from '../../helpers/api';
import '../Dashboard/Dashboard.css';
import './AdminDashboard.css';

const PERIOD_OPTIONS = [
  'today',
  'last_7_days',
  'last_30_days',
  'this_month',
  'last_month',
  'this_year',
];

const TASK_TYPE_KEYS = {
  channel_model_test: 'channel_model_test',
  channel_refresh_models: 'channel_refresh_models',
  channel_refresh_balance: 'channel_refresh_balance',
};

const getQuotaPerUnit = () => {
  const raw = parseFloat(localStorage.getItem('quota_per_unit') || '1');
  if (!Number.isFinite(raw) || raw <= 0) return 1;
  return raw;
};

const toUsd = (quota) => {
  const value = Number(quota);
  if (!Number.isFinite(value)) return 0;
  return value / getQuotaPerUnit();
};

const formatUsd = (quota) => {
  const amount = toUsd(quota);
  if (!Number.isFinite(amount)) return '0.0000';
  return amount.toFixed(4);
};

const endOfDay = (date) => new Date(date.getFullYear(), date.getMonth(), date.getDate(), 23, 59, 59);

const startOfDay = (date) => new Date(date.getFullYear(), date.getMonth(), date.getDate(), 0, 0, 0);

const getPeriodRange = (period) => {
  const now = new Date();
  const todayStart = startOfDay(now);
  const todayEnd = endOfDay(now);
  let start = todayStart;
  let end = todayEnd;

  switch (period) {
    case 'today':
      start = todayStart;
      end = todayEnd;
      break;
    case 'last_7_days':
      start = new Date(todayStart);
      start.setDate(start.getDate() - 6);
      end = todayEnd;
      break;
    case 'last_30_days':
      start = new Date(todayStart);
      start.setDate(start.getDate() - 29);
      end = todayEnd;
      break;
    case 'this_month':
      start = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0);
      end = todayEnd;
      break;
    case 'last_month': {
      const currentMonthStart = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0);
      end = new Date(currentMonthStart.getTime() - 1000);
      start = new Date(end.getFullYear(), end.getMonth(), 1, 0, 0, 0);
      break;
    }
    case 'this_year':
      start = new Date(now.getFullYear(), 0, 1, 0, 0, 0);
      end = todayEnd;
      break;
    default:
      start = new Date(todayStart);
      start.setDate(start.getDate() - 29);
      end = todayEnd;
      break;
  }

  return {
    startTimestamp: Math.floor(start.getTime() / 1000),
    endTimestamp: Math.floor(end.getTime() / 1000),
  };
};

const statusColor = (status) => {
  switch (status) {
    case 1:
      return 'green';
    case 2:
      return 'grey';
    case 3:
      return 'orange';
    case 4:
      return 'blue';
    default:
      return 'grey';
  }
};

const taskStatusColor = (status) => {
  switch (status) {
    case 'pending':
      return 'yellow';
    case 'running':
      return 'blue';
    case 'succeeded':
      return 'green';
    case 'failed':
      return 'red';
    case 'canceled':
      return 'grey';
    default:
      return 'grey';
  }
};

const AdminDashboard = () => {
  const { t } = useTranslation();
  const [period, setPeriod] = useState('last_7_days');
  const [loading, setLoading] = useState(false);
  const [summary, setSummary] = useState({
    consumeQuota: 0,
    topupQuota: 0,
    netQuota: 0,
    channelTotal: 0,
    channelEnabled: 0,
    channelDisabled: 0,
    groupTotal: 0,
    providerTotal: 0,
    taskActiveTotal: 0,
    taskFailedTotal: 0,
  });
  const [topChannels, setTopChannels] = useState([]);
  const [recentTasks, setRecentTasks] = useState([]);
  const [lastUpdatedAt, setLastUpdatedAt] = useState(0);

  const periodOptions = useMemo(
    () =>
      PERIOD_OPTIONS.map((value) => ({
        key: value,
        value,
        text: t(`dashboard.spending.period.${value}`),
      })),
    [t]
  );

  const getQuotaStat = useCallback(async (type, startTimestamp, endTimestamp) => {
    const res = await API.get('/api/v1/admin/log/stat', {
      params: {
        type,
        start_timestamp: startTimestamp,
        end_timestamp: endTimestamp,
      },
    });
    if (!res.data?.success) return 0;
    return Number(res.data?.data?.quota || 0);
  }, []);

  const getPageTotal = useCallback(async (url, params = {}) => {
    const res = await API.get(url, {
      params: {
        page: 1,
        page_size: 1,
        ...params,
      },
    });
    if (!res.data?.success) return 0;
    const payload = res.data?.data;
    if (!payload || typeof payload !== 'object') return 0;
    return Number(payload.total || 0);
  }, []);

  const getAllChannels = useCallback(async () => {
    const pageSize = 100;
    let page = 1;
    let expectedTotal = 0;
    const rows = [];

    while (page <= 100) {
      const res = await API.get('/api/v1/admin/channels/', {
        params: {
          page,
          page_size: pageSize,
        },
      });
      if (!res.data?.success) break;
      const payload = res.data?.data;
      if (Array.isArray(payload)) {
        return payload;
      }
      const items = Array.isArray(payload?.items) ? payload.items : [];
      expectedTotal = Number(payload?.total || 0);
      if (items.length === 0) break;
      rows.push(...items);
      if ((expectedTotal > 0 && rows.length >= expectedTotal) || items.length < pageSize) {
        break;
      }
      page += 1;
    }

    return rows;
  }, []);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const { startTimestamp, endTimestamp } = getPeriodRange(period);
      const [
        consumeQuota,
        topupQuota,
        channels,
        groupTotal,
        providerTotal,
        taskActiveTotal,
        taskFailedTotal,
        recentTasksRes,
      ] = await Promise.all([
        getQuotaStat(2, startTimestamp, endTimestamp),
        getQuotaStat(1, startTimestamp, endTimestamp),
        getAllChannels(),
        getPageTotal('/api/v1/admin/groups/'),
        getPageTotal('/api/v1/admin/providers/'),
        getPageTotal('/api/v1/admin/tasks/', { status: 'pending,running' }),
        getPageTotal('/api/v1/admin/tasks/', { status: 'failed' }),
        API.get('/api/v1/admin/tasks/', { params: { page: 1, page_size: 8 } }),
      ]);

      const enabledCount = channels.filter((item) => Number(item.status) === 1).length;
      const disabledCount = channels.filter((item) => [2, 3].includes(Number(item.status))).length;
      const sortedChannels = [...channels].sort(
        (a, b) => Number(b.used_quota || 0) - Number(a.used_quota || 0)
      );

      setSummary({
        consumeQuota,
        topupQuota,
        netQuota: topupQuota - consumeQuota,
        channelTotal: channels.length,
        channelEnabled: enabledCount,
        channelDisabled: disabledCount,
        groupTotal,
        providerTotal,
        taskActiveTotal,
        taskFailedTotal,
      });
      setTopChannels(sortedChannels.slice(0, 8));
      setRecentTasks(Array.isArray(recentTasksRes.data?.data?.items) ? recentTasksRes.data.data.items : []);
      setLastUpdatedAt(Date.now());
    } catch (error) {
      console.error('Failed to load admin dashboard:', error);
      setSummary({
        consumeQuota: 0,
        topupQuota: 0,
        netQuota: 0,
        channelTotal: 0,
        channelEnabled: 0,
        channelDisabled: 0,
        groupTotal: 0,
        providerTotal: 0,
        taskActiveTotal: 0,
        taskFailedTotal: 0,
      });
      setTopChannels([]);
      setRecentTasks([]);
    } finally {
      setLoading(false);
    }
  }, [getAllChannels, getPageTotal, getQuotaStat, period]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const formatUpdatedAt = (value) => {
    if (!value) return '-';
    return new Date(value).toLocaleString('zh-CN', { hour12: false });
  };

  const renderCapabilities = (raw) => {
    if (!Array.isArray(raw) || raw.length === 0) return '-';
    return raw
      .map((item) => t(`dashboard.admin.capabilities.${item}`, { defaultValue: item }))
      .join(' / ');
  };

  return (
    <div className='dashboard-container admin-dashboard-container'>
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header className='router-card-header router-section-title'>
            {t('dashboard.admin.title')}
          </Card.Header>
          <div className='admin-dashboard-toolbar'>
            <div className='admin-dashboard-period'>
              <span className='admin-dashboard-period-label'>
                {t('dashboard.admin.period.label')}
              </span>
              <Dropdown
                className='router-section-dropdown'
                selection
                options={periodOptions}
                value={period}
                onChange={(e, { value }) => setPeriod(value)}
              />
            </div>
            <div className='admin-dashboard-toolbar-right'>
              <span className='admin-dashboard-updated'>
                {t('dashboard.admin.updated_at', { time: formatUpdatedAt(lastUpdatedAt) })}
              </span>
              <Button
                className='router-inline-button'
                type='button'
                loading={loading}
                onClick={loadData}
              >
                {t('dashboard.admin.buttons.refresh')}
              </Button>
            </div>
          </div>
          <div className='admin-dashboard-kpi-grid'>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.consume')}</div>
              <div className='admin-dashboard-kpi-value'>{formatUsd(summary.consumeQuota)}</div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.topup')}</div>
              <div className='admin-dashboard-kpi-value'>{formatUsd(summary.topupQuota)}</div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.net')}</div>
              <div className='admin-dashboard-kpi-value'>{formatUsd(summary.netQuota)}</div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.channels')}</div>
              <div className='admin-dashboard-kpi-value'>
                {summary.channelEnabled} / {summary.channelTotal}
              </div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.channel_disabled')}</div>
              <div className='admin-dashboard-kpi-value'>{summary.channelDisabled}</div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.groups')}</div>
              <div className='admin-dashboard-kpi-value'>{summary.groupTotal}</div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.providers')}</div>
              <div className='admin-dashboard-kpi-value'>{summary.providerTotal}</div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.tasks_active')}</div>
              <div className='admin-dashboard-kpi-value'>{summary.taskActiveTotal}</div>
            </div>
            <div className='admin-dashboard-kpi-item'>
              <div className='admin-dashboard-kpi-label'>{t('dashboard.admin.metrics.tasks_failed')}</div>
              <div className='admin-dashboard-kpi-value'>{summary.taskFailedTotal}</div>
            </div>
          </div>
        </Card.Content>
      </Card>

      <Card fluid className='chart-card admin-dashboard-section'>
        <Card.Content>
          <Card.Header className='router-card-header router-section-title'>
            {t('dashboard.admin.sections.channels')}
          </Card.Header>
          {topChannels.length === 0 ? (
            <div className='admin-dashboard-empty'>{t('dashboard.admin.empty.channels')}</div>
          ) : (
            <Table compact='very' basic='very' celled>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>{t('dashboard.admin.table.channel_name')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.status')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.capabilities')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.balance')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.used_cost')}</Table.HeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {topChannels.map((row) => (
                  <Table.Row key={row.id}>
                    <Table.Cell>{row.name || '-'}</Table.Cell>
                    <Table.Cell>
                      <Label size='tiny' color={statusColor(Number(row.status))}>
                        {t(`dashboard.admin.channel_status.${Number(row.status)}`, {
                          defaultValue: t('dashboard.admin.channel_status.default'),
                        })}
                      </Label>
                    </Table.Cell>
                    <Table.Cell>{renderCapabilities(row.capabilities)}</Table.Cell>
                    <Table.Cell>{Number(row.balance || 0).toFixed(4)}</Table.Cell>
                    <Table.Cell>{formatUsd(Number(row.used_quota || 0))}</Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          )}
        </Card.Content>
      </Card>

      <Card fluid className='chart-card admin-dashboard-section'>
        <Card.Content>
          <Card.Header className='router-card-header router-section-title'>
            {t('dashboard.admin.sections.tasks')}
          </Card.Header>
          {recentTasks.length === 0 ? (
            <div className='admin-dashboard-empty'>{t('dashboard.admin.empty.tasks')}</div>
          ) : (
            <Table compact='very' basic='very' celled>
              <Table.Header>
                <Table.Row>
                  <Table.HeaderCell>{t('dashboard.admin.table.task_type')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.task_status')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.task_channel')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.task_model')}</Table.HeaderCell>
                  <Table.HeaderCell>{t('dashboard.admin.table.task_created')}</Table.HeaderCell>
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {recentTasks.map((task) => (
                  <Table.Row key={task.id}>
                    <Table.Cell>
                      {t(`dashboard.admin.task_type.${TASK_TYPE_KEYS[task.type] || 'default'}`, {
                        defaultValue: task.type || '-',
                      })}
                    </Table.Cell>
                    <Table.Cell>
                      <Label size='tiny' color={taskStatusColor(task.status)}>
                        {t(`dashboard.admin.task_status.${task.status || 'default'}`, {
                          defaultValue: task.status || '-',
                        })}
                      </Label>
                    </Table.Cell>
                    <Table.Cell>{task.channel_name || '-'}</Table.Cell>
                    <Table.Cell>{task.model || '-'}</Table.Cell>
                    <Table.Cell>{task.created_at ? formatUpdatedAt(task.created_at * 1000) : '-'}</Table.Cell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default AdminDashboard;
