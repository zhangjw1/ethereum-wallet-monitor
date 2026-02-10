import React, { useEffect, useState } from 'react';
import styles from './TokenAnalysis.module.css';
import { Shield, AlertTriangle, CheckCircle, XCircle, Loader2, RefreshCw, Search } from 'lucide-react';
import { getTokens, type TokenAnalysis as TokenAnalysisType } from '../api';

function shortAddr(addr: string): string {
    if (!addr) return '-';
    if (addr.length <= 12) return addr;
    return addr.slice(0, 6) + '...' + addr.slice(-4);
}

function riskColor(level: string): string {
    switch (level) {
        case 'low': return 'var(--success)';
        case 'medium': return 'var(--warning)';
        case 'high': return 'var(--danger)';
        case 'critical': return '#7C3AED';
        default: return 'var(--text-secondary)';
    }
}

function riskBg(level: string): string {
    switch (level) {
        case 'low': return '#ECFDF5';
        case 'medium': return '#FFFBEB';
        case 'high': return '#FEF2F2';
        case 'critical': return '#F5F3FF';
        default: return '#F1F5F9';
    }
}

function statusLabel(s: string): string {
    const map: Record<string, string> = {
        'PENDING_LIQUIDITY': '待加池',
        'ANALYZING': '分析中',
        'MONITORING': '监控中',
        'POTENTIAL': '潜力币',
        'REJECTED': '已拒绝',
        'RUGGED': '已跑路',
        'EXPIRED': '已过期',
    };
    return map[s] || s;
}

const TokenAnalysis: React.FC = () => {
    const [tokens, setTokens] = useState<TokenAnalysisType[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [filter, setFilter] = useState('');
    const [searchAddr, setSearchAddr] = useState('');

    const fetchData = async (opts?: { status?: string; address?: string }) => {
        setLoading(true);
        setError('');
        try {
            const params: any = { limit: 30 };
            if (opts?.status) params.status = opts.status;
            if (opts?.address) params.address = opts.address;
            const data = await getTokens(params);
            setTokens(Array.isArray(data) ? data : [data as TokenAnalysisType]);
        } catch (e: any) {
            setError(e.message || '加载失败');
            setTokens([]);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
    }, []);

    const handleFilterChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        const val = e.target.value;
        setFilter(val);
        fetchData(val ? { status: val } : undefined);
    };

    const handleSearch = () => {
        if (searchAddr.trim()) {
            fetchData({ address: searchAddr.trim() });
        } else {
            fetchData(filter ? { status: filter } : undefined);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter') handleSearch();
    };

    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <div>
                    <h2>代币分析</h2>
                    <p className={styles.subtitle}>追踪新代币安全评估、风险评分与流动性状态。</p>
                </div>
                <div className={styles.controls}>
                    <select className={styles.select} value={filter} onChange={handleFilterChange}>
                        <option value="">全部状态</option>
                        <option value="MONITORING">监控中</option>
                        <option value="ANALYZING">分析中</option>
                        <option value="PENDING_LIQUIDITY">待加池</option>
                        <option value="POTENTIAL">潜力币</option>
                        <option value="REJECTED">已拒绝</option>
                        <option value="RUGGED">已跑路</option>
                    </select>
                    <button className="btn btn-outline" onClick={() => { setSearchAddr(''); setFilter(''); fetchData(); }}>
                        <RefreshCw size={16} />
                    </button>
                </div>
            </div>

            <div className={styles.toolbar}>
                <div className={styles.searchBox}>
                    <Search size={16} />
                    <input
                        className={styles.searchInput}
                        placeholder="按代币合约地址搜索..."
                        value={searchAddr}
                        onChange={e => setSearchAddr(e.target.value)}
                        onKeyDown={handleKeyDown}
                    />
                    <button className="btn btn-primary" onClick={handleSearch}>搜索</button>
                </div>
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
                    <button className="btn btn-outline" onClick={() => fetchData()}>重试</button>
                </div>
            )}

            {!loading && !error && tokens.length === 0 && (
                <div className={styles.state}>暂无代币分析数据</div>
            )}

            <div className={`card ${styles.tableCard}`}>
                <table className={styles.table}>
                    <thead>
                        <tr>
                            <th>代币</th>
                            <th>合约地址</th>
                            <th>风险评分</th>
                            <th>蜜罐</th>
                            <th>买入税 / 卖出税</th>
                            <th>流动性</th>
                            <th>状态</th>
                        </tr>
                    </thead>
                    <tbody>
                        {tokens.map((t) => (
                            <tr key={t.id}>
                                <td>
                                    <div className={styles.tokenName}>
                                        <strong>{t.symbol || '-'}</strong>
                                        <span className={styles.dimText}>{t.name || ''}</span>
                                    </div>
                                </td>
                                <td>
                                    <a
                                        className={styles.addrLink}
                                        href={`https://etherscan.io/address/${t.token_address}`}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        title={t.token_address}
                                    >
                                        {shortAddr(t.token_address)}
                                    </a>
                                </td>
                                <td>
                                    <span
                                        className={styles.riskBadge}
                                        style={{ background: riskBg(t.risk_level), color: riskColor(t.risk_level) }}
                                    >
                                        <Shield size={12} />
                                        {t.risk_score.toFixed(0)} · {t.risk_level || '-'}
                                    </span>
                                </td>
                                <td>
                                    {t.is_honeypot
                                        ? <span className={styles.honeypotYes}><XCircle size={14} /> 是</span>
                                        : <span className={styles.honeypotNo}><CheckCircle size={14} /> 否</span>
                                    }
                                </td>
                                <td className={styles.taxCell}>
                                    {t.buy_tax.toFixed(1)}% / {t.sell_tax.toFixed(1)}%
                                </td>
                                <td>
                                    {t.has_liquidity
                                        ? <span className={styles.liqYes}>${t.liquidity_usd.toLocaleString()}</span>
                                        : <span className={styles.dimText}>无</span>
                                    }
                                </td>
                                <td>
                                    <span className={styles.statusBadge}>
                                        {statusLabel(t.status)}
                                    </span>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default TokenAnalysis;
