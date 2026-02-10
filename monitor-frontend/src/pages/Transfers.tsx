import React, { useEffect, useState } from 'react';
import styles from './Transfers.module.css';
import { ArrowUpRight, ArrowDownLeft, Search, RefreshCw, Loader2 } from 'lucide-react';
import { getTransferRecords, type TransferRecord } from '../api';

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

const Transfers: React.FC = () => {
    const [transfers, setTransfers] = useState<TransferRecord[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [searchAddr, setSearchAddr] = useState('');

    const fetchData = async (address?: string) => {
        setLoading(true);
        setError('');
        try {
            const params: any = { limit: 30 };
            if (address) params.address = address;
            const data = await getTransferRecords(params);
            setTransfers(Array.isArray(data) ? data : [data]);
        } catch (e: any) {
            setError(e.message || '加载失败');
            setTransfers([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleSearch = () => {
        fetchData(searchAddr.trim() || undefined);
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') handleSearch();
    };

    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <div className={styles.pageInfo}>
                    <h2>交易流水</h2>
                    <p>实时监控钱包转账记录，支持按地址搜索。</p>
                </div>
                <div className={styles.stats}>
                    <div className={styles.statItem}>
                        <span className={styles.statLabel}>记录数</span>
                        <span className={styles.statValue}>{transfers.length}</span>
                    </div>
                </div>
            </div>

            <div className={styles.toolbar}>
                <div className={styles.searchBox}>
                    <Search size={16} />
                    <input
                        className={styles.searchInput}
                        placeholder="按钱包地址搜索..."
                        value={searchAddr}
                        onChange={e => setSearchAddr(e.target.value)}
                        onKeyDown={handleKeyDown}
                    />
                    <button className="btn btn-primary" onClick={handleSearch}>搜索</button>
                </div>
                <button className="btn btn-outline" onClick={() => { setSearchAddr(''); fetchData(); }}>
                    <RefreshCw size={16} /> 刷新
                </button>
            </div>

            {loading && (
                <div className={styles.loadingState}>
                    <Loader2 size={24} className={styles.spin} />
                    <span>加载中...</span>
                </div>
            )}

            {error && (
                <div className={styles.errorState}>
                    <span>{error}</span>
                    <button className="btn btn-outline" onClick={() => fetchData()}>重试</button>
                </div>
            )}

            {!loading && !error && transfers.length === 0 && (
                <div className={styles.emptyState}>暂无交易流水数据</div>
            )}

            <div className={styles.feed}>
                {transfers.map((tx) => (
                    <div key={tx.id} className={styles.logItem}>
                        <div className={`${styles.iconBase} ${tx.direction === '转入' ? styles.iconIn : styles.iconOut}`}>
                            {tx.direction === '转入' ? <ArrowDownLeft size={20} /> : <ArrowUpRight size={20} />}
                        </div>
                        <div className={styles.content}>
                            <div className={styles.rowTop}>
                                <span className={styles.addresses}>
                                    <span className={styles.addr} title={tx.from_address}>{shortAddr(tx.from_address)}</span>
                                    <span className={styles.arrow}>→</span>
                                    <span className={styles.addr} title={tx.to_address}>{shortAddr(tx.to_address)}</span>
                                </span>
                                <span className={styles.amount}>{tx.amount} {tx.currency}</span>
                            </div>
                            <div className={styles.rowBottom}>
                                <span className={styles.time}>
                                    {tx.monitor_label && <span className={styles.label}>{tx.monitor_label}</span>}
                                    {timeAgo(tx.created_at)} · Block #{tx.block_number}
                                </span>
                                <a
                                    className={styles.txLink}
                                    href={`https://etherscan.io/tx/${tx.tx_hash}`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    title={tx.tx_hash}
                                >
                                    {shortAddr(tx.tx_hash)}
                                </a>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default Transfers;
