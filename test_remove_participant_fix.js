const API_BASE_URL = 'http://localhost:8083/api/v1';

// Test data
const TEST_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTAwOTE0OTIsImlhdCI6MTc1MDA4Nzg5Miwic3ViIjoiZmQ0MzRjMGUtOTVkZS00MWQwLWE1NzYtOWQ0ZWEyZmVkN2U5IiwidXNlcl9pZCI6ImZkNDM0YzBlLTk1ZGUtNDFkMC1hNTc2LTlkNGVhMmZlZDdlOSJ9.SdUmJZK0aSpJdHiXk9PfD903fNj24uCdAtwR2bnlHDw';
const TEST_USER_ID = 'fd434c0e-95de-41d0-a576-9d4ea2fed7e9';

console.log('Testing Remove Participant Fix...');

// Test 1: Create a test group chat with participants
async function createTestChatWithParticipants() {
    console.log('\n1. Creating test group chat with multiple participants...');
    
    try {
        const response = await fetch(`${API_BASE_URL}/chats`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${TEST_TOKEN}`
            },
            body: JSON.stringify({
                type: 'group',
                name: 'Remove Test Group Chat',
                participants: [TEST_USER_ID]
            })
        });

        const data = await response.json();
        console.log('Create chat response:', response.status, data);
        
        if (response.ok && data.chat) {
            const chatId = data.chat.id;
            
            // Add a test participant to remove later
            console.log('\n   Adding test participant to remove...');
            const addResponse = await fetch(`${API_BASE_URL}/chats/${chatId}/participants`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${TEST_TOKEN}`
                },
                body: JSON.stringify({
                    user_id: 'b1234567-89ab-cdef-0123-456789abcdef'
                })
            });
            
            const addData = await addResponse.json();
            console.log('   Add participant response:', addResponse.status, addData);
            
            return chatId;
        } else if (response.ok && data.chat_id) {
            return data.chat_id;
        }
        
        console.error('Failed to create test chat:', data);
        return null;
    } catch (error) {
        console.error('Error creating test chat:', error);
        return null;
    }
}

// Test 2: List participants with detailed structure analysis
async function analyzeParticipantStructure(chatId) {
    console.log(`\n2. Analyzing participant data structure in chat ${chatId}...`);
    
    try {
        const response = await fetch(`${API_BASE_URL}/chats/${chatId}/participants`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${TEST_TOKEN}`
            }
        });

        const data = await response.json();
        console.log('List participants response:', response.status);
        console.log('Full response structure:', JSON.stringify(data, null, 2));
        
        if (response.ok) {
            let participants = [];
            if (data.data && data.data.participants) {
                participants = data.data.participants;
            } else if (data.participants) {
                participants = data.participants;
            }
            
            console.log('\n   Participant structure analysis:');
            participants.forEach((p, index) => {
                console.log(`   Participant ${index + 1}:`);
                console.log(`     - id: ${p.id || 'undefined'}`);
                console.log(`     - user_id: ${p.user_id || 'undefined'}`);
                console.log(`     - username: ${p.username || 'undefined'}`);
                console.log(`     - is_admin: ${p.is_admin || false}`);
                console.log(`     - All fields:`, Object.keys(p));
                console.log('');
            });
            
            return participants;
        } else {
            console.log('❌ Failed to list participants:', data.message || data);
            return [];
        }
    } catch (error) {
        console.error('❌ Error listing participants:', error);
        return [];
    }
}

// Test 3: Test remove with proper ID handling
async function testRemoveWithProperID(chatId, participants) {
    console.log(`\n3. Testing remove participant with proper ID handling...`);
    
    if (participants.length < 2) {
        console.log('❌ Need at least 2 participants to test removal');
        return false;
    }
    
    // Find a non-admin participant to remove
    const participantToRemove = participants.find(p => !p.is_admin);
    if (!participantToRemove) {
        console.log('❌ No non-admin participant found to remove');
        return false;
    }
    
    // Test the ID resolution logic that the frontend now uses
    const userIdToRemove = participantToRemove.id || participantToRemove.user_id;
    
    console.log(`   Removing participant with ID: ${userIdToRemove}`);
    console.log(`   (resolved from: id=${participantToRemove.id}, user_id=${participantToRemove.user_id})`);
    
    try {
        const response = await fetch(`${API_BASE_URL}/chats/${chatId}/participants/${userIdToRemove}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${TEST_TOKEN}`
            }
        });

        const data = await response.json();
        console.log('   Remove participant response:', response.status, data);
        
        if (response.ok) {
            console.log('✅ Participant removed successfully!');
            return true;
        } else {
            console.log('❌ Failed to remove participant:', data.message || data);
            return false;
        }
    } catch (error) {
        console.error('❌ Error removing participant:', error);
        return false;
    }
}

// Test 4: Verify removal by listing participants again
async function verifyRemoval(chatId) {
    console.log(`\n4. Verifying participant removal...`);
    
    const participants = await analyzeParticipantStructure(chatId);
    
    if (participants.length === 1) {
        console.log('✅ Removal verified - only 1 participant remains (the admin)');
        return true;
    } else {
        console.log(`❌ Unexpected participant count: ${participants.length}`);
        return false;
    }
}

// Run all tests
async function runRemoveParticipantTests() {
    console.log('🚀 Starting remove participant fix tests...');
    
    // Create test chat with participants
    const chatId = await createTestChatWithParticipants();
    if (!chatId) {
        console.log('❌ Cannot continue tests without a valid chat ID');
        return;
    }
    
    // Analyze participant structure
    const participants = await analyzeParticipantStructure(chatId);
    
    // Test removal with proper ID handling
    const removeSuccess = await testRemoveWithProperID(chatId, participants);
    
    // Verify removal
    if (removeSuccess) {
        await verifyRemoval(chatId);
    }
    
    console.log('\n🏁 Remove participant tests completed!');
}

// For running in Node.js
if (typeof require !== 'undefined' && require.main === module) {
    runRemoveParticipantTests();
}

// For running in browser
if (typeof window !== 'undefined') {
    window.runRemoveParticipantTests = runRemoveParticipantTests;
}
