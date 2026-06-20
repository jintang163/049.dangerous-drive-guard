import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
export default defineConfig(function (_a) {
    var mode = _a.mode;
    var env = loadEnv(mode, process.cwd(), '');
    var isDashboard = mode === 'dashboard';
    return {
        plugins: [react()],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, './src'),
            },
        },
        server: {
            port: 3000,
            open: false,
            proxy: {
                '/api': {
                    target: env.VITE_API_BASE || 'http://localhost:8888',
                    changeOrigin: true,
                },
                '/ws': {
                    target: env.VITE_WS_BASE || 'ws://localhost:8888',
                    ws: true,
                    changeOrigin: true,
                },
            },
        },
        build: {
            outDir: isDashboard ? 'dist-dashboard' : 'dist-admin',
            sourcemap: mode !== 'production',
            rollupOptions: {
                output: {
                    manualChunks: {
                        'react-vendor': ['react', 'react-dom', 'react-router-dom', 'recoil'],
                        'antd-vendor': ['antd', '@ant-design/icons', '@ant-design/pro-components'],
                        'charts-vendor': ['echarts', 'echarts-for-react', '@ant-design/plots'],
                        'map-vendor': ['@amap/amap-jsapi-loader'],
                    },
                },
            },
            chunkSizeWarningLimit: 1500,
        },
        css: {
            preprocessorOptions: {
                less: {
                    modifyVars: {
                        'primary-color': '#1677ff',
                        'border-radius-base': '6px',
                    },
                    javascriptEnabled: true,
                },
            },
        },
    };
});
