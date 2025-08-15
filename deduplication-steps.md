# Deduplication Implementation Plan

## Overview
Add concurrent duplicate detection to prevent processing articles with identical titles that are already stored in the vector database. This will save API quota by skipping duplicates and demonstrate practical concurrent programming patterns.

## Current Architecture Analysis

### Key Components
- **Worker Pool**: 10 concurrent workers processing RSS items
- **Processing Pipeline**: RSS → Summarize → Embed → Store
- **Vector Database**: Weaviate with title-based search capability
- **Quota Management**: First 50 items only

### Current Flow
1. Parse RSS feed and get all items
2. Limit to first 50 items
3. Send jobs to worker pool
4. Each worker: summarize → embed → store
5. Track success/failure counts

## Implementation Steps

### Step 1: Add Title Search Method to WeaviateDB
**File**: `vectordb/weaviate.go`
- Add `SearchByTitle(ctx context.Context, title string) (*Article, error)` method
- Use GraphQL query to search for exact title matches
- Return first match or nil if not found

### Step 2: Create Deduplication Service
**File**: `deduplication/deduplicator.go` (new)
- Create `Deduplicator` struct with vector DB reference
- Add `IsDuplicate(ctx context.Context, title string) (bool, error)` method
- Handle concurrent access safely

### Step 3: Modify Worker Function
**File**: `main.go` - `worker()` function
- Add deduplication check as first step in worker
- If duplicate found: log skip message and return early
- If not duplicate: proceed with normal processing
- Update result tracking to distinguish skipped vs processed

### Step 4: Update Result Tracking
**File**: `main.go` - `ProcessingResult` struct and main processing
- Add `Skipped bool` field to `ProcessingResult`
- Update result collection to track: processed, skipped, failed
- Modify logging to show all three categories

### Step 5: Enhance Quota Logic
**File**: `main.go` - main function
- Change from "first 50 items" to "first 50 unique items"
- Continue processing until 50 unique articles are processed
- Stop early if RSS feed is exhausted

## Concurrent Programming Concepts Demonstrated

### 1. Race Condition Prevention
- Multiple workers checking for duplicates simultaneously
- Potential race: Worker A checks title, Worker B checks same title, both proceed
- Solution: Database-level uniqueness or worker coordination

### 2. Worker Pool Pattern Enhancement
- Workers now have early exit path (deduplication)
- Demonstrates conditional processing in concurrent systems
- Shows how to handle variable processing times

### 3. Channel Communication
- Results channel now carries more information (skipped vs processed)
- Demonstrates structured communication between goroutines

## Expected Behavior Changes

### Before Implementation
- Always processes exactly 50 items (or fewer if RSS has less)
- May store duplicate articles from different RSS sources
- Uses full API quota even for duplicates

### After Implementation
- Processes items until 50 unique articles are stored
- Skips duplicates with fast database lookup
- Saves API quota by avoiding duplicate processing
- Shows concurrent duplicate detection in action

## Files to Modify

1. `vectordb/weaviate.go` - Add title search method
2. `main.go` - Modify worker and result tracking
3. `deduplication/deduplicator.go` - New service (optional, can inline)

## Key Metrics to Show

- **Before**: "Processed 50/50 items"
- **After**: "Processed 35/50 unique items, skipped 15 duplicates"
- **API Savings**: Reduced OpenAI API calls by ~30%
- **Performance**: Faster processing due to early exits