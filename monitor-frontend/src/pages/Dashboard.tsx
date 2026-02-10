import React, { useEffect, useState } from 'react';
import KPICard from '../components/KPICard';
import CodePreview from '../components/CodePreview';
import styles from './Dashboard.module.css';
import { getNotifications, getTransferRecords, type NotificationStats, type TransferRecord } from '../api';

const Dashboard: React.FC = () => {
    const [stats, setStats] = useState<NotificationStats | null>(null);
    const [latestTx, setLatestTx] = useState<TransferRecord | null>(null);
    const [streamData, setStreamData] = useState<any>({});
    const [transferCount, setTransferCount] = useState(0);

    // 加载通知统计
    const loadStats = async () => {
        try {
            const data = await getNotifications({ stats: true });
            setStats(data as NotificationStats);
        } catch { /* ignore */ }
    };

    // 加载最新流水
    const loadLatest = async () => {
        try {
            const data = await getTransferRecords({ limit: 5 });
            const list = Array.isArray(data) ? data : [data];
            setTransferCount(list.length);
            if (list.length > 0) {
                setLatestTx(list[0] as TransferRecord);
                updateStream(list[0] as TransferRecord);
            }
        } catch { /* ignore */ }
    };

    const updateStream = (tx: TransferRecord) => {
        setStreamData({
            event: tx.direction === '转入' ? 'Transfer In' : 'Transfer Out',
            currency: tx.currency,
            amount: tx.amount,
            from: tx.from_address,
            to: tx.to_address,
            block: tx.block_number,
            txHash: tx.tx_hash,
            timestamp: tx.created_at,
        });
    };

    useEffect(() => {
        loadStats();
        loadLatest();

        // 轮询每 15 秒更新
        const interval = setInterval(() => {
            loadStats();
            loadLatest();
        }, 15000);
        return () => clearInterval(interval);
    }, []);

    // 按类型统计主要金额
    const typeBreakdown = stats?.by_type
        ? stats.by_type.map(t => `${t.Type}: ${t.Count}`).join(', ')
        : '-';

    return (
        <div className={styles.container}>
            <section className={styles.hero}>
                <div className={styles.heroContent}>
                    <h2 className={styles.heroTitle}>系统运行中</h2>
                    <p className={styles.heroSubtitle}>
                        以太坊主网钱包监控 · 实时推送通知 · 代币安全分析
                    </p>
                </div>
                <div className={styles.heroActions}>
                    <button className="btn btn-primary" onClick={() => { loadStats(); loadLatest(); }}>刷新数据</button>
                </div>
            </section>

            <section className={styles.stats}>
                <KPICard
                    title="通知总数"
                    value={stats ? String(stats.total) : '-'}
                    change={stats ? `今日 ${stats.today}` : ''}
                    isPositive={true}
                />
                <KPICard
                    title="发送成功"
                    value={stats ? String(stats.success) : '-'}
                    change={stats && stats.total > 0 ? `${((stats.success / stats.total) * 100).toFixed(1)}%` : ''}
                    isPositive={true}
                />
                <KPICard
                    title="发送失败"
                    value={stats ? String(stats.failed) : '-'}
                    change={stats && stats.total > 0 ? `${((stats.failed / stats.total) * 100).toFixed(1)}%` : ''}
                    isPositive={false}
                />
                <KPICard
                    title="最近交易"
                    value={String(transferCount)}
                    change={latestTx ? `${latestTx.currency} ${latestTx.direction}` : ''}
                    isPositive={true}
                />
            </section>

            <section className={styles.integration}>
                <div className={styles.integrationInfo}>
                    <h3 className={styles.sectionTitle}>实时数据流</h3>
                    <p className={styles.sectionText}>
                        展示最新的一笔钱包监控交易流水，每 15 秒自动刷新。
                    </p>
                    {stats?.by_type && stats.by_type.length > 0 && (
                        <div className={styles.apiCodes}>
                            <code>通知分布: {typeBreakdown}</code>
                        </div>
                    )}
                </div>
                <div className={styles.previewContainer}>
                    <CodePreview
                        title="Latest Transfer (Live)"
                        data={Object.keys(streamData).length > 0 ? streamData : { message: '暂无数据' }}
                    />
                </div>
            </section>
        </div>
    );
};

export default Dashboard;
