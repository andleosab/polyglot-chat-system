package org.demo.chat;

import java.util.Set;

import io.quarkus.websockets.next.UserData.TypedKey;

public final class WsKeys {

	//TODO: revisit - check if UUID is used currently and rename it 
	public static final TypedKey<String> USERNAME = TypedKey.forString("username");
	public static final TypedKey<Set<Long>> GROUPS = new TypedKey<Set<Long>>("groups");

}
