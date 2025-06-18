#!/bin/bash

echo "🔍 Testing Google Cloud Configuration..."
echo "========================================"

# Check if credentials file exists
if [ ! -f "credentials/google-cloud-key.json" ]; then
    echo "❌ Error: Google Cloud credentials file not found!"
    echo "   Please place your credentials file at: credentials/google-cloud-key.json"
    exit 1
fi

echo "✅ Credentials file found"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found!"
    echo "   Please create .env file with your Google Cloud configuration"
    exit 1
fi

echo "✅ .env file found"

# Load environment variables
source .env

# Check required environment variables
if [ -z "$GOOGLE_CLOUD_PROJECT_ID" ]; then
    echo "❌ Error: GOOGLE_CLOUD_PROJECT_ID not set in .env"
    exit 1
fi

if [ -z "$GOOGLE_CLOUD_LOCATION" ]; then
    echo "❌ Error: GOOGLE_CLOUD_LOCATION not set in .env"
    exit 1
fi

if [ -z "$GOOGLE_CLOUD_CATALOG" ]; then
    echo "❌ Error: GOOGLE_CLOUD_CATALOG not set in .env"
    exit 1
fi

echo "✅ Environment variables configured:"
echo "   Project ID: $GOOGLE_CLOUD_PROJECT_ID"
echo "   Location: $GOOGLE_CLOUD_LOCATION"
echo "   Catalog: $GOOGLE_CLOUD_CATALOG"

# Test Google Cloud authentication
echo ""
echo "🔐 Testing Google Cloud authentication..."
export GOOGLE_APPLICATION_CREDENTIALS="$GOOGLE_CLOUD_CREDENTIALS_FILE"

# Test if gcloud is available and authenticated
if command -v gcloud &> /dev/null; then
    echo "✅ gcloud CLI found"
    
    # Test authentication
    if gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        echo "✅ Google Cloud authentication successful"
    else
        echo "⚠️  Warning: No active gcloud authentication found"
        echo "   This is OK if you're using service account credentials"
    fi
else
    echo "⚠️  Warning: gcloud CLI not found"
    echo "   This is OK if you're using service account credentials directly"
fi

echo ""
echo "🎯 Configuration Summary:"
echo "========================="
echo "✅ Credentials file: $GOOGLE_CLOUD_CREDENTIALS_FILE"
echo "✅ Project ID: $GOOGLE_CLOUD_PROJECT_ID"
echo "✅ Location: $GOOGLE_CLOUD_LOCATION"
echo "✅ Catalog: $GOOGLE_CLOUD_CATALOG"
echo ""
echo "🚀 Ready to use Retail Search!"
echo ""
echo "Next steps:"
echo "1. Make sure your Retail API is enabled in Google Cloud Console"
echo "2. Create a catalog in Vertex AI > Retail > Catalogs"
echo "3. Run the application: go run main.go"
echo "4. Test with: curl 'http://localhost:8080/api/v1/products/search?q=test&lat=-23.550520&lng=-46.633308&radius=5.0'" 