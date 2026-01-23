import Alpine from 'alpinejs';
import htmx from 'htmx.org';
import { initFlashListeners, showFlashMessage } from './flash.js';
import { copyToClipboard } from './clipboard.js';
import { initDashboardCharts, updateChart, registerDashboardData, initAnalyticsCharts, initAdminOverviewCharts } from './charts.js';
import { registerQrCodeModal } from './qrcode.js';

// Expose HTMX to window
window.htmx = htmx;

// Initialize Flash Listeners
initFlashListeners();

// Register Alpine Components
registerDashboardData(Alpine);
registerQrCodeModal(Alpine);

// Start Alpine
window.Alpine = Alpine;
Alpine.start();

// Initialize Charts on DOMContentLoaded
document.addEventListener('DOMContentLoaded', () => {
    initDashboardCharts();
    initAnalyticsCharts();
    initAdminOverviewCharts();
});

// Re-init charts on HTMX swaps
document.addEventListener('htmx:afterSwap', (evt) => {
    // Check key elements to decide what to init
    if (evt.target.querySelector('#mainChart')) initDashboardCharts();
    if (evt.target.querySelector('#dailyUniqueVisitorsChart')) initAnalyticsCharts();
    if (evt.target.querySelector('#userGrowthChart')) initAdminOverviewCharts();
});

// Event Delegation for data-actions
document.addEventListener('click', (event) => {
    const actionTrigger = event.target.closest('[data-action]');
    if (!actionTrigger) return;

    const action = actionTrigger.getAttribute('data-action');
    const value = actionTrigger.getAttribute('data-value');

    switch (action) {
        case 'copy':
            if (value) {
                copyToClipboard(value);
                event.preventDefault();
                event.stopPropagation();
            }
            break;
    }
});

// Global exposure for legacy compatibility
window.updateChart = updateChart;
window.copyToClipboard = copyToClipboard;
window.showFlashMessage = showFlashMessage;
