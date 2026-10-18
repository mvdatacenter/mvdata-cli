const { CreateBucketCommand, S3Client } = require('@aws-sdk/client-s3')
const openbao = require('./test/openbao/server.js')

// The bucket name console-api writes audit records to: app/db/repos/secretAudit.ts.
const AUDIT_BUCKET = 'mvdata-console-audit'

// Each token holds its process's one policy, as a Kubernetes login to that role would issue, so a
// refusal in a test is pyx applying the policy file.
const TOKENS = {
    'pyx-reconciler': process.env.PYX_RECONCILER_TOKEN,
    'pyx-tenant-admin': process.env.PYX_TENANT_ADMIN_TOKEN,
}

async function ensureBucket() {
    const client = new S3Client({
        endpoint: process.env.SECRET_AUDIT_BUCKET_ENDPOINT,
        region: 'auto',
        forcePathStyle: true,
        credentials: {
            accessKeyId: process.env.SECRET_AUDIT_BUCKET_ACCESS_KEY_ID,
            secretAccessKey: process.env.SECRET_AUDIT_BUCKET_SECRET_ACCESS_KEY,
        },
    })
    // Each run starts an empty gateway, and the console's healthy wait outlasts the gateway's start.
    await client.send(new CreateBucketCommand({ Bucket: AUDIT_BUCKET }))
}

async function main() {
    for (const [policy, id] of Object.entries(TOKENS)) {
        if (!id) throw new Error(`no token id for ${policy}; run.sh sets PYX_RECONCILER_TOKEN and PYX_TENANT_ADMIN_TOKEN`)
    }
    // ready() also refuses any OpenBao but the version test/openbao/pin.json pins.
    await openbao.ready()
    await openbao.configureLikePyx()
    for (const [policy, id] of Object.entries(TOKENS)) {
        await openbao.call('POST', 'auth/token/create', { id, policies: [policy], no_default_policy: true })
    }
    await ensureBucket()
    console.log(`OpenBao configured like pyx with policies ${openbao.policyNames().join(', ')}; audit bucket ${AUDIT_BUCKET} ready`)
}

main().catch(err => {
    console.error(err)
    process.exit(1)
})
