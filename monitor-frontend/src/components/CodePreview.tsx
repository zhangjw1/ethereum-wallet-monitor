import React from 'react';
import styles from './CodePreview.module.css';

interface CodePreviewProps {
    title: string;
    data: any;
}

const CodePreview: React.FC<CodePreviewProps> = ({ title, data }) => {
    return (
        <div className={styles.container}>
            <div className={styles.header}>
                <span className={styles.title}>{title}</span>
                <span className={styles.liveBadge}>
                    <span className={styles.pulse}></span>
                    LIVE
                </span>
            </div>
            <div className={styles.window}>
                <pre className={styles.code}>
                    {JSON.stringify(data, null, 2)}
                </pre>
            </div>
        </div>
    );
};

export default CodePreview;
