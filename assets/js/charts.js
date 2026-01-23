import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns';
import jsVectorMap from 'jsvectormap';
import 'jsvectormap/dist/maps/world.js';
import 'jsvectormap/dist/jsvectormap.css';

function load2dData(name) {
    const element = document.getElementById(name);
    if (!element) return { labels: [], data: [] };
    
    const input = JSON.parse(element.textContent);
    const labels = [];
    const data = [];
    
    input.forEach(item => {
        // Handle different data structures
        if (item.Time) labels.push(item.Time);
        else if (item.Label) labels.push(item.Label);
        
        if (item.Count !== undefined) data.push(item.Count);
        else if (item.Value !== undefined) data.push(item.Value);
    });

    return { labels, data };
}

// Internal chart creation helpers
function createLineChart(ctx, labels, datasets, options = {}) {
    const defaultOptions = {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            y: { beginAtZero: true }
        }
    };

    return new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: datasets
        },
        options: { ...defaultOptions, ...options }
    });
}

function createDoughnutChart(ctx, labels, datasets, options = {}) {
    const defaultOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: { position: 'bottom' }
        }
    };

    return new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: labels,
            datasets: datasets
        },
        options: { ...defaultOptions, ...options }
    });
}

function getData() {
    const element = document.getElementById('visitOverTime');
    if (!element) return [];
    try {
        const input = JSON.parse(element.textContent);
        return input.map((p) => {
            return {
                "x": new Date(p.Time),
                "y": p.Count,
            }
        });
    } catch (e) {
        console.error("Failed to parse visitOverTime data", e);
        return [];
    }
}

let mainChart = null;

export function updateChart() {
    if (!mainChart) return;
    const data = getData();
    mainChart.data.datasets[0].data = data;
    mainChart.update();
}

function initMap() {
    const element = document.getElementById('locationData');
    if (!element) return;
    
    let locations = [];
    try {
        locations = JSON.parse(element.textContent) || [];
    } catch(e) {
        console.error("Failed to parse locationData", e);
        return;
    }

    const mapData = {};
    locations.forEach(l => {
        mapData[l.CountryCode] = l.VisitCount;
    });

    if (!document.getElementById('jvm-map')) return;

    new jsVectorMap({
        selector: '#jvm-map',
        map: 'world',
        visualizeData: {
            scale: ['#a5b4fc', '#4338ca'],
            values: mapData
        },
        zoomButtons: false,
        onRegionTooltipShow(event, tooltip, code) {
            tooltip.text(tooltip.text() + ' (' + (mapData[code] || 0) + ')');
        }
    });
}

export function initDashboardCharts() {
    initMap();

    const data = getData();
    const ctxMainElement = document.getElementById('mainChart');
    
    if (ctxMainElement) {
        const ctxMain = ctxMainElement.getContext('2d');
        mainChart = new Chart(ctxMain, {
            type: 'line',
            data: {
                datasets: [{
                    label: 'Clicks',
                    data: data,
                    borderColor: '#4f46e5',
                    backgroundColor: 'rgba(79, 70, 229, 0.05)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0,
                    pointHoverRadius: 6
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { display: false } },
                scales: {
                    y: { 
                        beginAtZero: true, 
                        grid: { borderDash: [2, 4], color: '#f1f5f9' },
                        ticks: { font: { size: 11 } }
                    },
                    x: { 
                        type: 'time',
                        time: {
                            unit: 'hour'
                        },
                        grid: { display: false },
                        ticks: { font: { size: 11 } }
                    }
                },
                interaction: {
                    intersect: false,
                    mode: 'index',
                },
            }
        });
    }

    const ctxRefElement = document.getElementById('referrerChart');
    if (ctxRefElement) {
        const {labels, data: refData} = load2dData('referrerChartData');
        const ctxRef = ctxRefElement.getContext('2d');
        new Chart(ctxRef, {
            type: 'doughnut',
            data: {
                labels: labels,
                datasets: [{
                    data: refData,
                    backgroundColor: ['#6366f1', '#a5b4fc', '#c7d2fe', '#e0e7ff', '#312e81'],
                    borderWidth: 0,
                    hoverOffset: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                cutout: '75%',
                plugins: { legend: { display: false } }
            }
        });
    }
}

export function initAnalyticsCharts() {
    const dailyVisitorsEl = document.getElementById('dailyUniqueVisitorsChart');
    if (dailyVisitorsEl) {
        const {labels: duvLabels, data: duvData} = load2dData('dailyUniqueVisitorsData');
        const duvCtx = dailyVisitorsEl.getContext('2d');
        
        const formattedDuvLabels = duvLabels.map(label => {
            const d = new Date(label);
            return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
        });

        createLineChart(duvCtx, formattedDuvLabels, [{
            label: 'Unique Visitors',
            data: duvData,
            backgroundColor: 'rgba(16, 185, 129, 0.2)',
            borderColor: 'rgba(16, 185, 129, 1)',
            borderWidth: 2,
            tension: 0.3,
            fill: true
        }]);
    }

    const clickDistEl = document.getElementById('clickDistributionChart');
    if (clickDistEl) {
        const {labels: cdLabels, data: cdData} = load2dData('clickDistributionData');
        const cdCtx = clickDistEl.getContext('2d');
        
        createDoughnutChart(cdCtx, cdLabels, [{
            data: cdData,
            backgroundColor: [
                'rgba(99, 102, 241, 0.8)',
                'rgba(139, 92, 246, 0.8)',
                'rgba(59, 130, 246, 0.8)',
                'rgba(16, 185, 129, 0.8)',
                'rgba(245, 158, 11, 0.8)'
            ],
            borderWidth: 1
        }]);
    }
}

export function initAdminOverviewCharts() {
    // User Growth Chart
    const userGrowthCtx = document.getElementById('userGrowthChart');
    if (userGrowthCtx) {
        var {labels, data } = load2dData('userGrowthChartData');
        var totalUserData = load2dData('totalUserChartData');
        
        new Chart(userGrowthCtx.getContext('2d'), {
            type: 'line',
            data: {
                labels: labels.map((label)=>{
                    let d = new Date(label);
                    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
                }),
                datasets: [{
                    label: 'New Users',
                    data: data,
                    backgroundColor: 'rgba(99, 102, 241, 0.2)',
                    borderColor: 'rgba(99, 102, 241, 1)',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true,
                    yAxisID: 'left'
                },
                {
                    label: 'Total Users',
                    data: totalUserData.data,
                    backgroundColor: 'rgba(24, 196, 157, 0.2)',
                    borderColor: 'rgba(24, 196, 157, 1)',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true,
                    yAxisID: 'right'
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    left: { position: 'left', beginAtZero: true },
                    right: { position: 'right', beginAtZero: true }
                }
            }
        });
    }

    // URL Trends Chart
    const urlTrendsCtx = document.getElementById('urlTrendsChart');
    if (urlTrendsCtx) {
        var {labels, data } = load2dData('urlTrendsChartData');
        new Chart(urlTrendsCtx.getContext('2d'), {
            type: 'bar',
            data: {
                labels: labels.map((label)=>{
                    let d = new Date(label);
                    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
                }),
                datasets: [{
                    label: 'URLs Created',
                    data: data,
                    backgroundColor: 'rgba(139, 92, 246, 0.8)',
                    borderColor: 'rgba(139, 92, 246, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: { y: { beginAtZero: true } }
            }
        });
    }

    // Daily Active Users Chart
    const dailyActiveUsersCtx = document.getElementById('dailyActiveUsersChart');
    if (dailyActiveUsersCtx) {
        new Chart(dailyActiveUsersCtx.getContext('2d'), {
            type: 'line',
            data: {
                labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
                datasets: [{
                    label: 'Active Users',
                    data: [3200, 3800, 4200, 4500, 4100, 2800, 2400],
                    backgroundColor: 'rgba(16, 185, 129, 0.2)',
                    borderColor: 'rgba(16, 185, 129, 1)',
                    borderWidth: 2,
                    tension: 0.3,
                    fill: true
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: { y: { beginAtZero: true } }
            }
        });
    }

    // Click Distribution Chart (Duplicate from initAnalyticsCharts, but might be different context)
    // Checking if element exists to avoid double init if called twice
    const clickDistributionCtx = document.getElementById('clickDistributionChart');
    // NOTE: If Analytics page also calls this, we might have conflict.
    // Ideally, we check if a chart instance already exists attached to the canvas.
    // But for now, simple check if element exists.
    // If this is the Overview page, it won't conflict with Analytics page unless they are SPA loaded together?
    // HTMX swaps body content often.
    if (clickDistributionCtx && !Chart.getChart(clickDistributionCtx)) {
         new Chart(clickDistributionCtx.getContext('2d'), {
            type: 'doughnut',
            data: {
                labels: ['Direct', 'Social Media', 'Search Engines', 'Email', 'Other'],
                datasets: [{
                    data: [35, 25, 20, 15, 5],
                    backgroundColor: [
                        'rgba(99, 102, 241, 0.8)',
                        'rgba(139, 92, 246, 0.8)',
                        'rgba(59, 130, 246, 0.8)',
                        'rgba(16, 185, 129, 0.8)',
                        'rgba(245, 158, 11, 0.8)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: { legend: { position: 'bottom' } }
            }
        });
    }
}

export function registerDashboardData(Alpine) {
    Alpine.data('dashboardData', () => ({
        timeRange: '24h',
    }));
}
