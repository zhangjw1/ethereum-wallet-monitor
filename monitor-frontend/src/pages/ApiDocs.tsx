import React from 'react';

const ApiDocs: React.FC = () => {
    return (
        <div style={{ maxWidth: '800px' }}>
            <h2>API Documentation</h2>
            <br />
            <div className="card">
                <h3>WebSocket Feed</h3>
                <p style={{ color: 'var(--text-secondary)', margin: '1rem 0' }}>
                    Connect to our WebSocket endpoint to receive real-time updates.
                </p>
                <div style={{ background: '#F1F5F9', padding: '1rem', borderRadius: '4px', fontFamily: 'var(--font-code)' }}>
                    wss://api.ethermonitor.io/v1/stream
                </div>
            </div>
        </div>
    );
};

export default ApiDocs;
