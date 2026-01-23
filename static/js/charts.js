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

function createBarChart(ctx, labels, datasets, options = {}) {
    const defaultOptions = {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
            y: { beginAtZero: true }
        }
    };

    return new Chart(ctx, {
        type: 'bar',
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

function createPieChart(ctx, labels, datasets, options = {}) {
    const defaultOptions = {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: { position: 'bottom' }
        }
    };

    return new Chart(ctx, {
        type: 'pie',
        data: {
            labels: labels,
            datasets: datasets
        },
        options: { ...defaultOptions, ...options }
    });
}

// Expose these to window so they can be called from HTMX or inline scripts
window.getData = function() {
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

window.updateChart = function() {
    if (!window.mainChart) return;
    const data = window.getData();
    window.mainChart.data.datasets[0].data = data;
    window.mainChart.update();
}

window.initMap = function() {
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

window.initDashboardCharts = function() {
    window.initMap();

    const data = window.getData();
    const ctxMainElement = document.getElementById('mainChart');
    
    if (ctxMainElement) {
        const ctxMain = ctxMainElement.getContext('2d');
        window.mainChart = new Chart(ctxMain, {
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
