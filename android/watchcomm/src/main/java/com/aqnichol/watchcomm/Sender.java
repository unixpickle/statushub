package com.aqnichol.watchcomm;

import android.content.Context;

import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.wearable.MessageApi;
import com.google.android.gms.wearable.Node;
import com.google.android.gms.wearable.NodeApi;
import com.google.android.gms.wearable.Wearable;

import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.concurrent.TimeUnit;

/**
 * Send messages from a background thread.
 */
public class Sender {
    private static final int CONNECTION_TIMEOUT = 5000;

    private Context context;
    private GoogleApiClient client;

    public Sender(Context ctx) {
        context = ctx;
    }

    public GoogleApiClient getClient() {
        return client;
    }

    public void connect() throws CommException {
        client = new GoogleApiClient.Builder(context)
                .addApi(Wearable.API)
                .build();
        client.blockingConnect(CONNECTION_TIMEOUT, TimeUnit.MILLISECONDS);
        if (!client.isConnected()) {
            client = null;
            throw new CommException("failed to connect to Play Services");
        }
    }

    public void disconnect() {
        if (client != null) {
            client.disconnect();
            client = null;
        }
    }

    public void sendMessage(String path, byte[] data) throws CommException {
        NodeApi.GetConnectedNodesResult result = Wearable.NodeApi.getConnectedNodes(client).await();
        if (!result.getStatus().isSuccess()) {
            throw new CommException("failed to list connected nodes");
        }
        List<Node> nodes = result.getNodes();
        if (nodes.size() > 0) {
            String id = nodes.get(0).getId();
            sendMessage(id, path, data);
        } else {
            throw new CommException("no connected nodes");
        }
    }

    public void sendMessage(String node, String path, byte[] data) throws CommException {
        MessageApi.SendMessageResult r = Wearable.MessageApi.sendMessage(client,
                node, path, data).await();
        if (!r.getStatus().isSuccess()) {
            throw new CommException("failed to send message");
        }
    }

    public void sendMessage(String node, String path, String data) throws CommException {
        sendMessage(node, path, data.getBytes(StandardCharsets.UTF_8));
    }
}
