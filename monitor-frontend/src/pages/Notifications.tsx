import React, { useEffect, useState } from 'react';
import styles from './Notifications.module.css';
import { Bell, CheckCircle, XCircle, Loader2, RefreshCw, Search, BarChart3 } from 'lucide-react';
import { getNotifications, type Notification as NotifType, type NotificationStats } from '../api';

function shortAddr(addr: string): string {
    if (!addr) return '-';
    if (addr.length <= 12) return addr;
    return addr.slice(0, 6) + '...' + addr.slice(-4);
}

function timeAgo(iso: string): string {
    const diff = Date.now() - new Date(iso).getTime();
    const s = Math.floor(diff / 1000);
    if (s < 60) return `${s}s ago`;
    const m = Math.floor(s / 60);
    if (m < 60) return `${m}m ago`;
    const h = Math.floor(m / 60);
    if (h < 24) return `${h}h ago`;
    return `${Math.floor(h / 24)}d ago`;
}

const Notifications: React.FC = () => {
    const [list, setList] = useState<NotifType[]>([]);
    const [stats, setStats] = useState<NotificationStats | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [typeFilter, setTypeFilter] = useState('');
    const [searchTx, setSearchTx] = useState('');

    const fetchList = async (type?: string) => {
        setLoading(true);
        setError('');
        try {
            const params: any = { limit: 30 };
            if (type) params.type = type;
            const data = await getNotifications(params);
            setList(Array.isArray(data) ? data : [data as NotifType]);
        } catch (e: any) {
            setError(e.message || '加载失败');
            setList([]);
        } finally {
            setLoading(false);
        }
    };

    const fetchStats = async () => {
        try {
            const data = await getNotifications({ stats: true });
            setStats(data as NotificationStats);
        } catch { /* ignore */ }
    };

    useEffect(() => {
        fetchList();
        fetchStats();
    }, []);

    const handleFilterChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        const val = e.target.value;
        setTypeFilter(val);
        fetchList(val || undefined);
    };

    const handleSearch = async () => {
        if (!searchTx.trim()) {
            fetchList(typeFilter || undefined);
            return;
        }
        setLoading(true);
        setError('');
        try {
            const data = await getNotifications({ tx_hash: searchTx.trim() });
            setList(Array.isArray(data) ? data : [data as NotifType]);
        } catch (e: any) {
            setError(e.message || '未找到');
            setList([]);
        } finally {
            setLoading(false);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') handleSearch();
    };

    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <div>
                    <h2>通知记录</h2>
                    <p className={styles.subtitle}>查看所有推送通知历史及统计。</p>
                </div>
                <button className="btn btn-outline" onClick={() => { setSearchTx(''); setTypeFilter(''); fetchList(); fetchStats(); }}>
                    <RefreshCw size={16} /> 刷新
                </button>
            </div>

            {/* Stats cards */}
            {stats && (
                <div className={styles.statsRow}>
                    <div className={styles.statCard}>
                        <BarChart3 size={18} className={styles.statIcon} />
                        <div>
                            <div className={styles.statNum}>{stats.total}</div>
                            <div className={styles.statLabel}>总通知</div>
                        </div>
                    </div>
                    <div className={styles.statCard}>
                        <CheckCircle size={18} style={{ color: 'var(--success)' }} />
                        <div>
                            <div className={styles.statNum}>{stats.success}</div>
                            <div className={styles.statLabel}>成功</div>
                        </div>
                    </div>
                    <div className={styles.statCard}>
                        <XCircle size={18} style={{ color: 'var(--danger)' }} />
                        <div>
                            <div className={styles.statNum}>{stats.failed}</div>
                            <div className={styles.statLabel}>失败</div>
                        </div>
                    </div>
                    <div className={styles.statCard}>
                        <Bell size={18} style={{ color: 'var(--primary)' }} />
                        <div>
                            <div className={styles.statNum}>{stats.today}</div>
                            <div className={styles.statLabel}>今日</div>
                        </div>
                    </div>
                </div>
            )}

            {/* Toolbar */}
            <div className={styles.toolbar}>
                <div className={styles.searchBox}>
                    <Search size={16} />
                    <input
                        className={styles.searchInput}
                        placeholder="按交易哈希搜索..."
                        value={searchTx}
                        onChange={e => setSearchTx(e.target.value)}
                        onKeyDown={handleKeyDown}
                    />
                    <button className="btn btn-primary" onClick={handleSearch}>搜索</button>
                </div>
                <select className={styles.select} value={typeFilter} onChange={handleFilterChange}>
                    <option value="">全部类型</option>
                    <option value="ETH_TRANSFER">ETH 转账</option>
                    <option value="USDT_TRANSFER">USDT 转账</option>
                    <option value="USDC_TRANSFER">USDC 转账</option>
                </select>
            </div>

            {loading && (
                <div className={styles.state}>
                    <Loader2 size={24} className={styles.spin} />
                    <span>加载中...</span>
                </div>
            )}

            {error && (
                <div className={`${styles.state} ${styles.stateError}`}>
                    <span>{error}</span>
                    <button className="btn btn-outline" onClick={() => fetchList()}>重试</button>
                </div>
            )}

            {!loading && !error && list.length === 0 && (
                <div className={styles.state}>暂无通知记录</div>
            )}

            {/* Table */}
            <div className={`card ${styles.tableCard}`}>
                <table className={styles.table}>
                    <thead>
                        <tr>
                            <th>时间</th>
                            <th>类型</th>
                            <th>方向</th>
                            <th>金额</th>
                            <th>From → To</th>
                            <th>状态</th>
                            <th>Tx Hash</th>
                        </tr>
                    </thead>
                    <tbody>
                        {list.map((n) => (
                            <tr key={n.id}>
                                <td className={styles.timeCell}>{timeAgo(n.created_at)}</td>
                                <td>
                                    <span className={styles.typeBadge}>{n.type}</span>
                                </td>
                                <td>{n.direction || '-'}</td>
                                <td className={styles.amountCell}>{n.amount} {n.currency}</td>
                                <td className={styles.addrCell}>
                                    <span title={n.from_address}>{shortAddr(n.from_address)}</span>
                                    <span className={styles.arrow}>→</span>
                                    <span title={n.to_address}>{shortAddr(n.to_address)}</span>
                                </td>
                                <td>
                                    {n.status === 'success'
                                        ? <span className="badge badge-success">成功</span>
                                        : <span className={styles.failBadge}>失败</span>
                                    }
                                </td>
                                <td>
                                    <a
                                        className={styles.txLink}
                                        href={`https://etherscan.io/tx/${n.tx_hash}`}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        title={n.tx_hash}
                                    >
                                        {shortAddr(n.tx_hash)}
                                    </a>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default Notifications;
