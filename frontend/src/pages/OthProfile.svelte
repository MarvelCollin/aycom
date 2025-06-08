        <div class="profile-name-container">
          <h1 class="profile-name">
            {profileData.name}
            {#if profileData.is_verified}
              <span class="user-verified-badge">
                <CheckCircleIcon size="18" />
              </span>
            {/if}
          </h1>
          <div class="profile-username">@{profileData.username}</div>
        </div> 