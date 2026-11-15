export default {
    name: 'AppMetrics',
    data() {
        return {
            loading: false,
            metrics: null,
            error: '',
            refreshedAt: null,
        }
    },
    computed: {
        results() {
            return this.metrics?.results || {}
        },
        system() {
            return this.results.system || {}
        },
        devices() {
            return this.results.devices || { total: 0, logged_in: 0, connected: 0 }
        },
        memory() {
            return this.system.memory || {}
        },
        folders() {
            return this.system.folders || {}
        },
        filesystem() {
            return this.system.filesystem || {}
        },
    },
    methods: {
        openModal() {
            $('#modalAppMetrics')
                .modal({
                    autofocus: false,
                    observeChanges: true,
                })
                .modal('show')
            this.loadMetrics()
        },
        async loadMetrics() {
            if (this.loading) return
            this.loading = true
            this.error = ''
            try {
                const response = await window.http.get('/app/metrics')
                this.metrics = response.data || null
                this.refreshedAt = new Date()
            } catch (error) {
                this.error = error.response?.data?.message || error.message || 'Failed to fetch system metrics'
                showErrorInfo(this.error)
            } finally {
                this.loading = false
            }
        },
        formatBytes(value) {
            if (value === null || value === undefined || Number.isNaN(Number(value))) return 'Unknown'
            const bytes = Math.max(0, Number(value))
            const units = ['B', 'KB', 'MB', 'GB', 'TB']
            let size = bytes
            let unit = 0
            while (size >= 1024 && unit < units.length - 1) {
                size /= 1024
                unit++
            }
            const digits = unit === 0 ? 0 : (size >= 100 ? 0 : 1)
            return `${size.toFixed(digits)} ${units[unit]}`
        },
        formatDuration(value) {
            const total = Math.max(0, Math.floor(Number(value) || 0))
            const days = Math.floor(total / 86400)
            const hours = Math.floor((total % 86400) / 3600)
            const minutes = Math.floor((total % 3600) / 60)
            const seconds = total % 60
            const parts = []
            if (days) parts.push(`${days}d`)
            if (hours || days) parts.push(`${hours}h`)
            if (minutes || hours || days) parts.push(`${minutes}m`)
            parts.push(`${seconds}s`)
            return parts.join(' ')
        },
        formatNumber(value, digits = 0) {
            if (value === null || value === undefined || Number.isNaN(Number(value))) return 'Unknown'
            return Number(value).toLocaleString(undefined, { maximumFractionDigits: digits })
        },
        formatDate(value) {
            if (!value) return 'Unknown'
            const date = new Date(value)
            if (Number.isNaN(date.getTime())) return String(value)
            return date.toLocaleString()
        },
    },
    template: `
        <div class="teal card" @click="openModal" style="cursor: pointer">
            <div class="content">
                <a class="ui blue right ribbon label">System</a>
                <div class="header">System Metrics</div>
                <div class="description">
                    View API uptime, devices, memory, storage and disk capacity.
                </div>
            </div>
        </div>

        <div class="ui large modal" id="modalAppMetrics">
            <i class="close icon"></i>
            <div class="header">
                <i class="chart line icon"></i>
                System Metrics
            </div>

            <div class="scrolling content">
                <div v-if="loading && !metrics" class="ui active centered inline loader"></div>

                <div v-if="error" class="ui negative message">
                    <div class="header">Metrics unavailable</div>
                    <p>[[ error ]]</p>
                </div>

                <template v-if="metrics">
                    <div class="ui three stackable cards">
                        <div class="card">
                            <div class="content">
                                <div class="header">API</div>
                                <div class="meta">Version and uptime</div>
                                <div class="description">
                                    <div class="ui relaxed list">
                                        <div class="item"><strong>Version:</strong> [[ results.version || 'Unknown' ]]</div>
                                        <div class="item"><strong>Uptime:</strong> [[ formatDuration(system.uptime_seconds) ]]</div>
                                        <div class="item"><strong>Started:</strong> [[ formatDate(system.started_at) ]]</div>
                                        <div class="item"><strong>Go:</strong> [[ system.go_version || 'Unknown' ]]</div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="card">
                            <div class="content">
                                <div class="header">Devices</div>
                                <div class="meta">Current device manager</div>
                                <div class="description">
                                    <div class="ui three mini statistics">
                                        <div class="statistic">
                                            <div class="value">[[ devices.total || 0 ]]</div>
                                            <div class="label">Total</div>
                                        </div>
                                        <div class="statistic">
                                            <div class="value">[[ devices.logged_in || 0 ]]</div>
                                            <div class="label">Logged in</div>
                                        </div>
                                        <div class="statistic">
                                            <div class="value">[[ devices.connected || 0 ]]</div>
                                            <div class="label">Connected</div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div class="card">
                            <div class="content">
                                <div class="header">Storage mode</div>
                                <div class="meta">Database persistence</div>
                                <div class="description">
                                    <div class="ui relaxed list">
                                        <div class="item"><strong>Session store:</strong> [[ results.session_store || 'Unknown' ]]</div>
                                        <div class="item">
                                            <strong>Chat storage:</strong>
                                            <span class="ui tiny label" :class="results.chat_storage_persistent ? 'green' : 'red'">
                                                [[ results.chat_storage_persistent ? 'Persistent' : 'Temporary' ]]
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <h4 class="ui dividing header"><i class="memory icon"></i> Memory and runtime</h4>
                    <div class="ui two column stackable grid">
                        <div class="column">
                            <table class="ui compact celled table">
                                <tbody>
                                    <tr><td>Heap allocated</td><td class="right aligned">[[ formatBytes(memory.heap_alloc_bytes) ]]</td></tr>
                                    <tr><td>Runtime reserved</td><td class="right aligned">[[ formatBytes(memory.runtime_sys_bytes) ]]</td></tr>
                                    <tr><td>Process RSS</td><td class="right aligned">[[ formatBytes(memory.process_rss_bytes) ]]</td></tr>
                                    <tr><td>Container usage</td><td class="right aligned">[[ formatBytes(memory.container_bytes) ]]</td></tr>
                                    <tr><td>Container limit</td><td class="right aligned">[[ formatBytes(memory.container_limit_bytes) ]]</td></tr>
                                </tbody>
                            </table>
                        </div>
                        <div class="column">
                            <table class="ui compact celled table">
                                <tbody>
                                    <tr><td>CPU time</td><td class="right aligned">[[ formatNumber(system.cpu_seconds, 2) ]] s</td></tr>
                                    <tr><td>Goroutines</td><td class="right aligned">[[ formatNumber(memory.goroutines) ]]</td></tr>
                                    <tr><td>GC cycles</td><td class="right aligned">[[ formatNumber(memory.gc_cycles) ]]</td></tr>
                                    <tr><td>Go version</td><td class="right aligned">[[ system.go_version || 'Unknown' ]]</td></tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <h4 class="ui dividing header"><i class="hdd icon"></i> Disk and folders</h4>
                    <table class="ui compact celled table">
                        <thead>
                            <tr><th>Location</th><th class="right aligned">Size</th><th class="right aligned">Files</th></tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td>Media</td>
                                <td class="right aligned">[[ formatBytes(folders.media?.bytes) ]]</td>
                                <td class="right aligned">[[ formatNumber(folders.media?.files) ]]</td>
                            </tr>
                            <tr>
                                <td>Storages</td>
                                <td class="right aligned">[[ formatBytes(folders.storages?.bytes) ]]</td>
                                <td class="right aligned">[[ formatNumber(folders.storages?.files) ]]</td>
                            </tr>
                            <tr>
                                <td>Filesystem total</td>
                                <td class="right aligned">[[ formatBytes(filesystem.total_bytes) ]]</td>
                                <td class="right aligned">—</td>
                            </tr>
                            <tr>
                                <td>Filesystem free</td>
                                <td class="right aligned">[[ formatBytes(filesystem.free_bytes) ]]</td>
                                <td class="right aligned">—</td>
                            </tr>
                        </tbody>
                    </table>

                    <div v-if="refreshedAt" class="ui tiny grey message">
                        Last refreshed: [[ refreshedAt.toLocaleString() ]]
                    </div>
                </template>
            </div>

            <div class="actions">
                <button class="ui button" @click="loadMetrics" :class="{loading: loading}" :disabled="loading">
                    <i class="sync alternate icon"></i>
                    Refresh
                </button>
                <button class="ui positive button">
                    <i class="check icon"></i>
                    Close
                </button>
            </div>
        </div>
    `
}
