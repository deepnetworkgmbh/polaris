$(function () {
  var clusterScoreChart = new Chart("clusterScoreChart", {
    type: 'doughnut',
    data: {
      labels: ["Passing", "Warnings", "Errors"],
      datasets: [{
        data: [
          polarisAuditData.ClusterSummary.Results.Totals.Successes,
          polarisAuditData.ClusterSummary.Results.Totals.Warnings,
          polarisAuditData.ClusterSummary.Results.Totals.Errors,
        ],
        backgroundColor: ['#8BD2DC', '#f26c21', '#a11f4c'],
      }]
    },
    options: {
      responsive: false,
      cutoutPercentage: 60,
      legend: {
        display: false,
      },
    }
  });

  var scanResultsChart = new Chart("scanResultsChart", {
    type: 'doughnut',
    data: {
      labels: [ "No Data", "Passing", "Warnings", "Errors"],
      datasets: [{
        data: [
          polarisAuditData.ScanResults.NoData,
          polarisAuditData.ScanResults.Successes,
          polarisAuditData.ScanResults.Warnings,
          polarisAuditData.ScanResults.Errors,
        ],
        backgroundColor: [ '#ACB7BF', '#8BD2DC', '#f26c21', '#a11f4c'],
      }]
    },
    options: {
      responsive: false,
      cutoutPercentage: 60,
      legend: {
        display: false,
      },
    }
  });
});
