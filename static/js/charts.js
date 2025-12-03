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
