const { MongoClient } = require('mongodb');

// Migration script untuk menambahkan field wallet security ke collection users
async function migrateWalletSecurity() {
  const uri = process.env.MONGODB_URI || 'mongodb://localhost:27017/vybes';
  const client = new MongoClient(uri);

  try {
    await client.connect();
    console.log('Connected to MongoDB');

    const db = client.db();
    const usersCollection = db.collection('users');

    // 1. Update existing users dengan field wallet security default
    console.log('Updating existing users with wallet security fields...');
    
    const updateResult = await usersCollection.updateMany(
      {
        $or: [
          { walletNonce: { $exists: false } },
          { lastWalletAccess: { $exists: false } },
          { failedLoginAttempts: { $exists: false } },
          { accountLockedUntil: { $exists: false } }
        ]
      },
      {
        $set: {
          walletNonce: 0,
          lastWalletAccess: null,
          failedLoginAttempts: 0,
          accountLockedUntil: null
        }
      }
    );

    console.log(`Updated ${updateResult.modifiedCount} users`);

    // 2. Create indexes untuk wallet sessions
    console.log('Creating indexes for wallet_sessions collection...');
    
    const walletSessionsCollection = db.collection('wallet_sessions');
    await walletSessionsCollection.createIndex({ sessionToken: 1 }, { unique: true });
    await walletSessionsCollection.createIndex({ userId: 1 });
    await walletSessionsCollection.createIndex({ expiresAt: 1 }, { expireAfterSeconds: 0 });

    // 3. Create indexes untuk wallet audit logs
    console.log('Creating indexes for wallet_audit_logs collection...');
    
    const walletAuditCollection = db.collection('wallet_audit_logs');
    await walletAuditCollection.createIndex({ userId: 1 });
    await walletAuditCollection.createIndex({ txHash: 1 });
    await walletAuditCollection.createIndex({ createdAt: -1 });
    await walletAuditCollection.createIndex({ action: 1 });
    await walletAuditCollection.createIndex({ status: 1 });

    // 4. Create TTL index untuk audit logs (optional - hapus setelah 1 tahun)
    await walletAuditCollection.createIndex(
      { createdAt: 1 },
      { expireAfterSeconds: 365 * 24 * 60 * 60 } // 1 year
    );

    console.log('Migration completed successfully!');

  } catch (error) {
    console.error('Migration failed:', error);
    throw error;
  } finally {
    await client.close();
  }
}

// Rollback function jika diperlukan
async function rollbackWalletSecurity() {
  const uri = process.env.MONGODB_URI || 'mongodb://localhost:27017/vybes';
  const client = new MongoClient(uri);

  try {
    await client.connect();
    console.log('Connected to MongoDB');

    const db = client.db();
    const usersCollection = db.collection('users');

    // Remove wallet security fields
    console.log('Removing wallet security fields from users...');
    
    const updateResult = await usersCollection.updateMany(
      {},
      {
        $unset: {
          walletNonce: "",
          lastWalletAccess: "",
          failedLoginAttempts: "",
          accountLockedUntil: ""
        }
      }
    );

    console.log(`Updated ${updateResult.modifiedCount} users`);

    // Drop collections
    console.log('Dropping wallet security collections...');
    await db.collection('wallet_sessions').drop();
    await db.collection('wallet_audit_logs').drop();

    console.log('Rollback completed successfully!');

  } catch (error) {
    console.error('Rollback failed:', error);
    throw error;
  } finally {
    await client.close();
  }
}

// Validation function untuk memastikan migration berhasil
async function validateMigration() {
  const uri = process.env.MONGODB_URI || 'mongodb://localhost:27017/vybes';
  const client = new MongoClient(uri);

  try {
    await client.connect();
    console.log('Validating migration...');

    const db = client.db();
    
    // Check if collections exist
    const collections = await db.listCollections().toArray();
    const collectionNames = collections.map(c => c.name);
    
    console.log('Available collections:', collectionNames);
    
    if (!collectionNames.includes('wallet_sessions')) {
      console.error('❌ wallet_sessions collection not found');
      return false;
    }
    
    if (!collectionNames.includes('wallet_audit_logs')) {
      console.error('❌ wallet_audit_logs collection not found');
      return false;
    }

    // Check if users have wallet security fields
    const usersCollection = db.collection('users');
    const userWithFields = await usersCollection.findOne({
      walletNonce: { $exists: true },
      lastWalletAccess: { $exists: true },
      failedLoginAttempts: { $exists: true },
      accountLockedUntil: { $exists: true }
    });

    if (!userWithFields) {
      console.error('❌ Users do not have wallet security fields');
      return false;
    }

    // Check indexes
    const walletSessionsIndexes = await db.collection('wallet_sessions').indexes();
    const walletAuditIndexes = await db.collection('wallet_audit_logs').indexes();

    console.log('wallet_sessions indexes:', walletSessionsIndexes.map(i => i.name));
    console.log('wallet_audit_logs indexes:', walletAuditIndexes.map(i => i.name));

    console.log('✅ Migration validation passed!');
    return true;

  } catch (error) {
    console.error('❌ Migration validation failed:', error);
    return false;
  } finally {
    await client.close();
  }
}

// Main execution
async function main() {
  const command = process.argv[2];

  switch (command) {
    case 'migrate':
      await migrateWalletSecurity();
      break;
    case 'rollback':
      await rollbackWalletSecurity();
      break;
    case 'validate':
      await validateMigration();
      break;
    default:
      console.log('Usage:');
      console.log('  node migrate_wallet_security.js migrate   - Run migration');
      console.log('  node migrate_wallet_security.js rollback  - Rollback migration');
      console.log('  node migrate_wallet_security.js validate  - Validate migration');
      process.exit(1);
  }
}

// Run if called directly
if (require.main === module) {
  main().catch(console.error);
}

module.exports = {
  migrateWalletSecurity,
  rollbackWalletSecurity,
  validateMigration
};