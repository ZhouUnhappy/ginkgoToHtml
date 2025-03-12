package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
)

// TestCase represents a single test case in the report
type TestCase struct {
	Title       string
	Status      string
	LogContent  string
	HasEnter    bool
	DisplayTime string
}

// TestReport represents the entire test report
type TestReport struct {
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	TestCases    []TestCase
	GeneratedAt  string
}

func main() {
	// Define command-line flags
	inputFile := flag.String("input", "", "Path to the test report log file")
	outputFile := flag.String("output", "", "Path to the output HTML file")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Please provide both input and output file paths")
		os.Exit(1)
	}

	// Read the input file
	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse the test report
	report := parseTestReport(file)

	// Generate the HTML report
	err = generateHTMLReport(report, *outputFile)
	if err != nil {
		fmt.Printf("Error generating HTML report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test report successfully converted to HTML: %s\n", *outputFile)
	fmt.Printf("Total: %d, Passed: %d, Failed: %d, Skipped: %d\n",
		report.TotalTests, report.PassedTests, report.FailedTests, report.SkippedTests)
}

func parseTestReport(file *os.File) TestReport {
	scanner := bufio.NewScanner(file)

	var report TestReport
	report.GeneratedAt = time.Now().Format("2006-01-02 15:04:05")

	var currentCase TestCase
	var inTestCase bool
	var buffer strings.Builder
	var titleLines []string
	var inSummarizing bool

	// Skip the first section (before the first separator)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "------------------------------") {
			inTestCase = true
			break
		}
	}

	// Process each test case
	for scanner.Scan() {
		line := scanner.Text()

		// Check if we're in the "Summarizing" section at the end
		if strings.Contains(line, "Summarizing") {
			inSummarizing = true
		}

		// If we hit a separator, we're starting a new test case
		if strings.Contains(line, "------------------------------") {
			// Save the previous test case if it exists
			if inTestCase && !inSummarizing {
				// Set the title from collected title lines
				if len(titleLines) > 0 {
					currentCase.Title = strings.Join(titleLines, ", ")
				}

				currentCase.LogContent = buffer.String()
				report.TestCases = append(report.TestCases, currentCase)

				// Update counters
				report.TotalTests++
				switch currentCase.Status {
				case "pass":
					report.PassedTests++
				case "fail":
					report.FailedTests++
				case "skip":
					report.SkippedTests++
				}
			}

			// Reset for the next test case
			currentCase = TestCase{}
			buffer.Reset()
			titleLines = nil
			inTestCase = true
			continue
		}

		// Skip processing if we're in the Summarizing section
		if inSummarizing {
			continue
		}

		// Add the line to the current test case's log content
		buffer.WriteString(line + "\n")

		// Check for status indicators
		if strings.Contains(line, "[SKIPPED]") {
			currentCase.Status = "skip"
		} else if strings.Contains(line, "[FAILED]") {
			currentCase.Status = "fail"
		}

		// Check for the Enter marker to determine the title
		if strings.Contains(line, "> Enter") {
			currentCase.HasEnter = true
		}

		// If we haven't found the Enter marker yet, collect non-indented lines as title components
		// Skip lines that contain a forward slash
		if !currentCase.HasEnter && line != "" && !strings.Contains(line, "[SKIPPED]") && !strings.Contains(line, "/") {
			titleLines = append(titleLines, line)
		}
	}

	// Add the last test case if there is one and not in Summarizing section
	if inTestCase && !inSummarizing {
		// Set the title from collected title lines
		if len(titleLines) > 0 {
			currentCase.Title = strings.Join(titleLines, ", ")
		}

		currentCase.LogContent = buffer.String()
		report.TestCases = append(report.TestCases, currentCase)

		// Update counters
		report.TotalTests++
		switch currentCase.Status {
		case "pass":
			report.PassedTests++
		case "fail":
			report.FailedTests++
		case "skip":
			report.SkippedTests++
		}
	}

	// Set default status to "pass" for test cases without an explicit status
	for i := range report.TestCases {
		if report.TestCases[i].Status == "" {
			report.TestCases[i].Status = "pass"
			// Increment the passed count for each test case that defaults to pass
			report.PassedTests++
		}
	}

	// Recalculate total tests based on the sum of pass, fail, and skip
	report.TotalTests = report.PassedTests + report.FailedTests + report.SkippedTests

	return report
}

func generateHTMLReport(report TestReport, outputFile string) error {
	// Define the HTML template
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        .header {
            background-color: #f8f9fa;
            padding: 10px 15px;
            border-radius: 5px;
            margin-bottom: 15px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .header h1 {
            font-size: 1.5em;
            margin: 0 0 5px 0;
        }
        .header p {
            margin: 0;
            font-size: 0.9em;
            color: #6c757d;
        }
        .test-result {
            display: inline-block;
            margin-left: 10px;
            padding: 3px 8px;
            border-radius: 3px;
            font-size: 0.8em;
            font-weight: bold;
            text-transform: uppercase;
            background-color: #f8d7da;
            color: #721c24;
        }
        .summary-stats {
            display: flex;
            margin-top: 8px;
            gap: 15px;
        }
        .stat-item {
            display: flex;
            align-items: center;
            font-size: 0.9em;
        }
        .stat-label {
            margin-right: 5px;
            font-weight: normal;
        }
        .stat-value {
            font-weight: bold;
        }
        .stat-total {
            color: #495057;
        }
        .stat-passed {
            color: #155724;
        }
        .stat-failed {
            color: #721c24;
        }
        .stat-skipped {
            color: #856404;
        }
        .test-case {
            margin-bottom: 20px;
            border: 1px solid #ddd;
            border-radius: 5px;
            overflow: hidden;
        }
        .test-header {
            padding: 10px 15px;
            cursor: pointer;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .test-title {
            font-weight: bold;
            flex: 1;
        }
        .test-status {
            padding: 3px 8px;
            border-radius: 3px;
            font-size: 0.8em;
            font-weight: bold;
            text-transform: uppercase;
        }
        .status-pass {
            background-color: #d4edda;
            color: #155724;
        }
        .status-fail {
            background-color: #f8d7da;
            color: #721c24;
        }
        .status-skip {
            background-color: #fff3cd;
            color: #856404;
        }
        .test-content {
            padding: 15px;
            background-color: #f8f9fa;
            border-top: 1px solid #ddd;
            display: none;
            white-space: pre-wrap;
            font-family: monospace;
            font-size: 0.9em;
            max-height: 500px;
            overflow-y: auto;
        }
        .test-case.expanded .test-content {
            display: block;
        }
        .footer {
            margin-top: 20px;
            text-align: center;
            color: #6c757d;
            font-size: 0.9em;
        }
        .toggle-all {
            margin-bottom: 20px;
            padding: 8px 16px;
            background-color: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        .toggle-all:hover {
            background-color: #0069d9;
        }
        .filter-buttons {
            margin-bottom: 20px;
            display: flex;
            gap: 10px;
        }
        .filter-button {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: bold;
        }
        .filter-all {
            background-color: #e9ecef;
        }
        .filter-passed {
            background-color: #d4edda;
            color: #155724;
        }
        .filter-failed {
            background-color: #f8d7da;
            color: #721c24;
        }
        .filter-skipped {
            background-color: #fff3cd;
            color: #856404;
        }
        .filter-button.active {
            box-shadow: 0 0 0 2px #007bff;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Test Report {{if gt .FailedTests 0}}<span class="test-result">Failed</span>{{end}}</h1>
        <p>Generated at: {{.GeneratedAt}}</p>
        <div class="summary-stats">
            <div class="stat-item stat-total">
                <span class="stat-label">Total:</span>
                <span class="stat-value">{{.TotalTests}}</span>
            </div>
            <div class="stat-item stat-passed">
                <span class="stat-label">Passed:</span>
                <span class="stat-value">{{.PassedTests}}</span>
            </div>
            <div class="stat-item stat-failed">
                <span class="stat-label">Failed:</span>
                <span class="stat-value">{{.FailedTests}}</span>
            </div>
            <div class="stat-item stat-skipped">
                <span class="stat-label">Skipped:</span>
                <span class="stat-value">{{.SkippedTests}}</span>
            </div>
        </div>
    </div>
    
    <div class="filter-buttons">
        <button class="filter-button filter-all active" data-filter="all">All Tests</button>
        <button class="filter-button filter-passed" data-filter="pass">Passed</button>
        <button class="filter-button filter-failed" data-filter="fail">Failed</button>
        <button class="filter-button filter-skipped" data-filter="skip">Skipped</button>
    </div>
    
    <button class="toggle-all" id="toggleAll">Expand All</button>
    
    <div class="test-cases">
        {{range .TestCases}}
        <div class="test-case" data-status="{{.Status}}">
            <div class="test-header">
                <div class="test-title">{{.Title}}</div>
                <div class="test-status status-{{.Status}}">{{.Status}}</div>
            </div>
            <div class="test-content">{{.LogContent}}</div>
        </div>
        {{end}}
    </div>
    
    <div class="footer">
        <p>Generated by GinkgoOutputToHTML</p>
    </div>
    
    <script>
        // Add click event to test headers to toggle content visibility
        document.querySelectorAll('.test-header').forEach(header => {
            header.addEventListener('click', () => {
                const testCase = header.parentElement;
                testCase.classList.toggle('expanded');
            });
        });
        
        // Add click event to toggle all button
        const toggleAllButton = document.getElementById('toggleAll');
        let allExpanded = false;
        
        toggleAllButton.addEventListener('click', () => {
            const visibleTestCases = document.querySelectorAll('.test-case:not([style*="display: none"])');
            allExpanded = !allExpanded;
            
            visibleTestCases.forEach(testCase => {
                if (allExpanded) {
                    testCase.classList.add('expanded');
                } else {
                    testCase.classList.remove('expanded');
                }
            });
            
            toggleAllButton.textContent = allExpanded ? 'Collapse All' : 'Expand All';
        });
        
        // Add filter functionality
        document.querySelectorAll('.filter-button').forEach(button => {
            button.addEventListener('click', () => {
                // Update active button
                document.querySelectorAll('.filter-button').forEach(btn => {
                    btn.classList.remove('active');
                });
                button.classList.add('active');
                
                // Filter test cases
                const filter = button.getAttribute('data-filter');
                document.querySelectorAll('.test-case').forEach(testCase => {
                    if (filter === 'all' || testCase.getAttribute('data-status') === filter) {
                        testCase.style.display = '';
                    } else {
                        testCase.style.display = 'none';
                    }
                });
                
                // Reset expand/collapse state
                allExpanded = false;
                toggleAllButton.textContent = 'Expand All';
            });
        });
    </script>
</body>
</html>`

	// Parse the template
	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return err
	}

	// Create the output file
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Execute the template with the report data
	err = t.Execute(out, report)
	if err != nil {
		return err
	}

	return nil
}
