# Bidprentjes Search API

A high-performance web application designed for searching and viewing memorial card (*bidprentjes*) records. This application leverages the speed of Go and the powerful indexing capabilities of the Bleve search engine to provide a seamless search experience across large datasets.

## Key Features

- **Advanced Search**: Full-text search with support for both fuzzy and exact matching across names (first name, prefix, last name), dates (birth, death, years), and locations.
- **Multilingual UI**: Built-in support for English, Dutch, and German, including localized date formats and search labels.
- **Photo & Scan Support**:
    - **Photo (Boolean)**: Indicates if a photo is available in the primary record.
    - **Scans (Numbered Links)**: Associative metadata linking records to multiple high-resolution images hosted on a CDN.
- **Hybrid Data Management**:
    - Load data from local CSV files for development.
    - Automatic backup and restoration of the search index using Google Cloud Storage (GCS).
    - Robust CSV processing with concurrent indexing for large datasets (e.g., 100,000+ records).
- **Responsive Design**: Web-based search interface styled with Bootstrap and accessible via mobile or desktop.

## Project Structure

- `models/`: Go struct definitions for data entities and JSON marshaling.
- `store/`: The core logic for Bleve indexing, GCS integration, and data retrieval.
- `handlers/`: Web handlers for processing search queries and rendering templates.
- `templates/`: HTML templates for the search interface.
- `scripts/`: Python tools for data generation and conversion.

## Setup & Installation

### Prerequisites
- [Go](https://golang.org/dl/) (1.21 or higher)
- [Python 3](https://www.python.org/downloads/) (for data generation scripts)

### Installation
1. Clone the repository.
2. Initialize and download Go dependencies:
   ```bash
   make deps
   ```

### Configuration
The application is configured via environment variables:
- `PORT`: The port on which the server will run (default: `8080`).
- `CDN_BASE_URL`: The base URL where your scan images are hosted (e.g., `https://cdn.example.com/`). The app automatically appends `.jpg` to scan IDs.
- `STORAGE_BUCKET`: (Optional) The name of the Google Cloud Storage bucket for index backups.

## Usage

### Generating Test Data
To create a local dataset for testing:
```bash
make generate-data
```
This will create `data/bidprentjes.csv` and `data/scans.csv` with sample records.

### Running the Application
To start the search engine locally:
```bash
# Default port 8080
make run

# Custom port and CDN
PORT=9090 CDN_BASE_URL=https://cdn.rikribbers.nl/ make run
```
Access the interface at `http://localhost:8080/search`.

### Testing
Run the Go test suite to verify indexing and data consistency:
```bash
make test
```

## Development Commands
- `make build`: Compile the application into a binary.
- `make clean`: Remove build artifacts and temporary CSV data.
- `make fmt`: Format the Go source code.

## License
This project is licensed under the MIT License - see the LICENSE file for details.
